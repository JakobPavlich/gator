# Blog Aggregator

## Installation guide

You will need to install Postgres and Go to run this program

You can install it with:
`go install github.com/JakobPavlich/gator`

## Config and database schema

You will need to set a config file and database schema

1. A config file will be need in the root directory, named '.gatorconfig.json'
   Use your database connection string.
   `{"[your connection string],"current_user_name":""}`
2. For the schema **goose** was used.
   Firstly form the root directory run:
   `cd sql/schema`
   and then:
   `goose postgres '[your connection string]' up`

## Usage

### Commands

- `gator register <user>`
- `gator login <user>`
- `gator addfeed <feed url>`
- `gator follow <feed name> <feed url>`
- `gator browse <?limit>`
  ...
