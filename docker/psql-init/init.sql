CREATE TABLE users (
    "user_id"   BIGINT NOT NULL,
    "name"      TEXT NOT NULL,
    "nick_name" TEXT,
    "summary"   TEXT,
    PRIMARY KEY ("user_id")
);

CREATE TABLE messages (
    "msg_id"   BIGINT NOT NULL,
    "user_id"  BIGINT NOT NULL,
    "group_id" BIGINT NOT NULL,
    "content"  TEXT NOT NULL,
    "raw"      TEXT,
    "deleted"  BOOLEAN NOT NULL DEFAULT FALSE,
    "is_cmd"   BOOLEAN NOT NULL DEFAULT FALSE,
    "time"     TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY ("msg_id"),
    FOREIGN KEY ("user_id") REFERENCES users(user_id)
);

CREATE TABLE group_llm_configs (
    "group_id"    BIGINT  NOT NULL,
    "prompt"      TEXT    NOT NULL,
    "max_history" INTEGER NOT NULL DEFAULT 200,
    "enabled"     BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY ("group_id")
);
