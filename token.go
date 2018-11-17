package apitoken

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/go-apibox/api"
	"github.com/go-apibox/utils"
	"github.com/go-xorm/xorm"
)

type Token struct {
	app      *api.App
	disabled bool
	inited   bool

	engine        *xorm.Engine
	tableName     string
	length        int
	actionMatcher *utils.Matcher
}

func NewToken(app *api.App) *Token {
	app.Error.RegisterGroupErrors("token", ErrorDefines)

	t := new(Token)
	t.app = app

	cfg := app.Config
	disabled := cfg.GetDefaultBool("apitoken.disabled", false)
	t.disabled = disabled
	if disabled {
		return t
	}

	t.init()
	return t
}

func (t *Token) init() {
	if t.inited {
		return
	}
	app := t.app
	cfg := app.Config
	dbType := cfg.GetDefaultString("apitoken.db_type", "mysql")
	dbAlias := cfg.GetDefaultString("apitoken.db_alias", "default")
	tableName := cfg.GetDefaultString("apitoken.table_name", "api_token")
	tokenLen := cfg.GetDefaultInt("apitoken.length", 32)
	actionWhitelist := cfg.GetDefaultStringArray("apitoken.actions.whitelist", []string{"*"})
	actionBlacklist := cfg.GetDefaultStringArray("apitoken.actions.blacklist", []string{})

	matcher := utils.NewMatcher()
	matcher.SetWhiteList(actionWhitelist)
	matcher.SetBlackList(actionBlacklist)

	var engine *xorm.Engine
	var err error

	if tableName == "" {
		goto return_token
	}

	switch dbType {
	case "mysql":
		if api.HasDriver("mysql") {
			engine, err = app.DB.GetMysql(dbAlias)
			if err != nil {
				app.Logger.Warning("(apitoken) can't get mysql with alias: %s, ignore apitoken db config.", dbAlias)
				goto return_token
			}
		} else {
			app.Logger.Warning("(apitoken) mysql driver is not ready, ignore apitoken db config.")
			goto return_token
		}
	case "sqlite3":
		if api.HasDriver("sqlite3") {
			engine, err = app.DB.GetSqlite3(dbAlias)
			if err != nil {
				app.Logger.Warning("(apitoken) can't get sqlite3 with alias: %s, ignore apitoken db config.", dbAlias)
				goto return_token
			}
		} else {
			app.Logger.Warning("(apitoken) sqlite3 driver is not ready, ignore apitoken db config.")
			goto return_token
		}
	}

	// try to create table
	if tableName != "" {
		if dbType == "mysql" {
			engine.Exec(fmt.Sprintf(MYSQL_SCHEMA_TOKEN, tableName))
		} else if dbType == "sqlite3" {
			sql := strings.Replace(SQLITE3_SCHEMA_TOKEN, "%s", tableName, -1)
			sqls := strings.Split(sql, ";")
			for _, s := range sqls {
				if s != "" {
					engine.Exec(s)
				}
			}
		}
	}

	// engine.ShowDebug = true

return_token:

	t.engine = engine
	t.tableName = tableName
	t.length = tokenLen
	t.actionMatcher = matcher
	t.inited = true
}

func (t *Token) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if t.disabled {
		next(w, r)
		return
	}

	c, err := api.NewContext(t.app, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// check if action not need to check token
	action := c.Input.GetAction()
	if !t.actionMatcher.Match(action) {
		next(w, r)
		return
	}

	token := c.Input.Get("api_token")
	if token == "" {
		api.WriteResponse(c, c.Error.NewGroupError("token", errorMissingToken))
		return
	}
	if len(token) != t.length {
		api.WriteResponse(c, c.Error.NewGroupError("token", errorInvalidToken))
		return
	}

	results, err := t.engine.Query("SELECT * FROM `"+t.tableName+"` WHERE `token_id`=?", token)
	if err != nil {
		api.WriteResponse(c, c.Error.New(api.ErrorInternalError, "QueryFailed").SetMessage(err.Error()))
		return
	}

	params := make(url.Values)
	for k, v := range r.Form {
		// 移除除appid外的所有全局参数
		if strings.HasPrefix(k, "api_") && k != "api_appid" {
			continue
		}
		params[k] = v
	}
	paramsStr := EncodeValues(params)
	paramsMd5 := fmt.Sprintf("%x", md5.Sum([]byte(paramsStr)))

	if len(results) > 0 {
		tAction := string(results[0]["action"])
		tParamMd5 := string(results[0]["params_md5"])
		if tAction != action || tParamMd5 != paramsMd5 {
			api.WriteResponse(c, c.Error.NewGroupError("token", errorParamMismatch))
			return
		}

		var data interface{}
		responseBody := results[0]["response_body"]
		json.Unmarshal(responseBody, &data)
		api.WriteResponse(c, data)
		return
	}

	// next middleware
	next(w, r)

	returnData := c.Get("returnData")
	returnDataBytes, err := json.Marshal(returnData)
	if err != nil {
		t.app.Logger.Error("(apitoken) Json marshal failed: %s", err.Error())
		return
	}

	_, err = t.engine.Exec("INSERT INTO `"+t.tableName+"` "+
		"(token_id, action, request_time, params, params_md5, response_body) "+
		"VALUES (?, ?, ?, ?, ?, ?)",
		token, action, utils.Timestamp(), paramsStr, paramsMd5, string(returnDataBytes),
	)
	if err != nil {
		t.app.Logger.Error("(apitoken) Insert token failed: %s", err.Error())
		return
	}
}

// EncodeValues encodes the values into "URL encoded" form
// ("bar=baz&foo=quux") sorted by key and values.
// Changed from url.Values.Encode()
func EncodeValues(v url.Values) string {
	if v == nil {
		return ""
	}
	var buf bytes.Buffer
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]

		// we add this line
		sort.Strings(vs)

		prefix := url.QueryEscape(k) + "="
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(prefix)
			buf.WriteString(url.QueryEscape(v))
		}
	}
	return buf.String()
}

// Enable enable the middle ware.
func (t *Token) Enable() {
	t.disabled = false
	t.init()
}

// Disable disable the middle ware.
func (t *Token) Disable() {
	t.disabled = true
}
