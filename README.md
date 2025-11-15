# bino

bino is a web application for managing patients in distributed wildlife rescues that have many independent facilities/shelters (referred to as "homes").

For now, bino uses Google Drive for actual patient journals, and users must have Google accounts.

## Prerequisites

* [go](https://go.dev/dl/) - programming language
* [postgresql](https://www.postgresql.org/download/) - database
* [just](https://github.com/casey/just) - command runner
* [sass](https://sass-lang.com/install/) - generates CSS from SCSS
* [sqlc](https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html) - generates Go code from SQL
* [templ](https://templ.guide/) - HTML templating system for Go

## Setup

This is only tested on Linux. Other Unixes and WSL should work too.

### Get the code

After you have installed all the prerequisites, clone the repository:

```sh
git clone git@github.com:jonathangjertsen/bino.git
```

### Config

Initialize your config

```sh
cp config.default.json config.json
```

Fields that need to be customized are marked with `<angle brackets>`.

### Database

Initialize the database:

```sh
just init_db
```

The database will be empty, but will be migrated up to the latest version when the application runs.

### Session key

Generate the key that is used to encrypt session cookies.

```sh
just session_key
```

TODO: generate session keys on-demand in the application.

### OAuth consent screen

You need to set up a project in Google Cloud Console and configure the [OAuth consent screen](https://developers.google.com/workspace/guides/configure-oauth-consent). This determines the `ClientID` field in the config.

Store the oauth credentials to `secret/oauth.json` (`secret/` is gitignore'd).

### Service account for Google Drive

Create a service account with access to the drive you are going to use.

Store the service account credentials to `secret/serviceaccount.json`.

Update the `DriveBase`, `JournalFolder`, `TemplateFile` and `ExtraJournalFolders` config fields with the respective IDs.

## Running

To build and run the application:

```sh
just run
```

There are also some other commands in the justfile if you want to just build or just
run sqlc etc.

The application starts up a webserver and runs a few periodic background jobs
that keep the search index in sync and delete stale data (expired sessions and invitations). 

Operations in Google Drive go through a task queue that has access to the service account. Other than that, there is hardly any concurrency at the application layer.

