package main

const strNotImplemented = "not implemented"
const strWrongTypeForIndexField = "wrong type for index/field"

const strUsage = "Usage: ./rediseen [start/help/version]"
const strHeader = "rediseen " + rediseenVersion

const strHelpDoc = "\n\n" + strUsage + "\n\n" +
	"Configuration Items (via environment variables):\n" +
	"- REDISEEN_REDIS_URI: URI of your Redis database, e.g. `redis://:@localhost:6379`\n" +
	"- REDISEEN_HOST: host of the service. Default port is 'localhost'\n" +
	"- REDISEEN_PORT: port of the service. Default port is 8000\n" +
	"- REDISEEN_DB_EXPOSED: Redis logical database(s) to expose. e.g., `0`, `0;3;9`, `0-9;15`, or `*`\n" +
	"- REDISEEN_KEY_PATTERN_EXPOSED: Regular expression pattern, " +
	"representing the name pattern of keys that you intend to expose\n" +
	"- REDISEEN_KEY_PATTERN_EXPOSE_ALL: If you intend to expose *all* your keys, " +
	"set `REDISEEN_KEY_PATTERN_EXPOSE_ALL` to `true`"

const strLogo = " _____            _  _   _____\n" +
	"|  __ \\          | |(_) / ____|\n" +
	"| |__) | ___   __| | _ | (___    ___   ___  _ __ \n" +
	"|  _  / / _ \\ / _` || | \\___ \\  / _ \\ / _ \\| '_ \\\n" +
	"| | \\ \\|  __/| (_| || | ____) ||  __/|  __/| | | |\n" +
	"|_|  \\_\\\\___| \\__,_||_||_____/  \\___| \\___||_| |_|"
