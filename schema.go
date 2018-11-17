package apitoken

const MYSQL_SCHEMA_TOKEN = "" +
	"CREATE TABLE IF NOT EXISTS `%s` (\n" +
	"  `token_id` VARCHAR(255) COLLATE utf8_bin NOT NULL COMMENT 'API TOKEN',\n" +
	"  `action` VARCHAR(255) NOT NULL COMMENT '接口名称',\n" +
	"  `request_time` INT UNSIGNED NOT NULL COMMENT '请求时间',\n" +
	"  `params` TEXT NOT NULL COMMENT '业务参数，URLENCODE格式',\n" +
	"  `params_md5` VARCHAR(255) NOT NULL COMMENT '业务参数MD5值',\n" +
	"  `response_body` LONGTEXT NOT NULL COMMENT '响应内容，不含HEADER，json格式',\n" +
	"  PRIMARY KEY (`token_id`))\n" +
	"ENGINE = InnoDB DEFAULT CHARSET=utf8"

const SQLITE3_SCHEMA_TOKEN = "" +
	"CREATE TABLE IF NOT EXISTS `%s` (\n" +
	"	`token_id`	TEXT NOT NULL PRIMARY KEY, -- API TOKEN\n" +
	"	`action`	TEXT NOT NULL, -- 接口名称\n" +
	"	`request_time`	INTEGER NOT NULL, -- 请求时间\n" +
	"	`params`	TEXT NOT NULL, -- 业务参数，URLENCODE格式\n" +
	"	`params_md5`	TEXT NOT NULL, -- 业务参数MD5值\n" +
	"	`response_body`	TEXT NOT NULL -- 响应内容，不含HEADER，json格式\n" +
	");\n"
