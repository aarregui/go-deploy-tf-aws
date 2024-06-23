CREATE TABLE "accounts" (
  "id" uuid PRIMARY KEY NOT NULL,
  "email" varchar(250) UNIQUE NOT NULL,
  "password" char(60) NOT NULL,
  "active" boolean NOT NULL DEFAULT (false),
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp
);