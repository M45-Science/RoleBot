# RoleBot

RoleBot is a Discord self-service role bot. Members can add or remove roles
through a slash-command menu; members with **Manage Roles** can choose which
roles are available for self-assignment.

## Requirements

- Go 1.25 or newer
- A Discord application and bot token
- The bot invited with the `applications.commands` and `bot` scopes, plus the
  **Manage Roles** permission

The bot can only manage roles below its own highest role in Discord's role
hierarchy.

## Configure and run

On its first start, RoleBot creates `config.json`. Stop it, then add the bot
token, application ID, and an optional status message:

```json
{
  "Token": "your-bot-token",
  "App": "your-application-id",
  "Domain": "Self-service roles"
}
```

Register the slash commands once (and after changing their definitions):

```sh
go run . -regCommands
```

Run normally afterward:

```sh
go run .
```

Use `/configure-bot` to select self-assignable roles. Members use
`/get-roles` to add or remove their selections. To remove the registered
commands, run `go run . -deregCommands`.

## Data and compatibility

Runtime state is stored in `data/` and is intentionally ignored by Git.
The RoleBot rename does not change the existing database filenames or compact
timestamp epoch, so a deployment can retain its current `data/` directory.

## Development

```sh
go test ./...
go vet ./...
```

RoleBot is licensed under the Mozilla Public License 2.0; see [LICENSE](LICENSE).
