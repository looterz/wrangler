# wrangler
Unreal Engine 4 Dedicated Server manager built around Amazon Web Services.

## features
* Supports both steampipe and S3 for updating the server
* Automatic updates triggered from SNS notifications
* Automatic process monitoring and crash recovery
* Automatic server configuration via EC2 instance meta-data and tags

## config
The configuration file lives next to `wrangler.exe` and should be named `wrangler.toml`.

```toml
# The absolute path to steamcmd.exe
steamcmd = ""

# The absolute path to the servers root directory
server = ""

# The process name to monitor, typically the unreal process wrapper in the games root folder
process = ""

# Use Amazon S3 instead of steamcmd, requires that the instance have an assigned IAM role with S3 access
UseS3Bucket = false

# The bucket name
S3Bucket = ""

# The bucket prefix (not working)
S3BucketPrefix = ""

# The absolute path of the folder to download the server to
S3Folder = ""
```

## tags
Wrangler reads and uses these tags when configuring itself and the server process.

* `Server_Branch` - The steampipe branch to pull updates from, if `live` it is ommitted and uses the default branch in steampipe
* `Server_Name` - The name to assign to the server when launching the game server, game needs to be configured to parse and use this and treat underscores as spaces in `GameSession::RegisterServer`
* `Server_Map` - The map argument passed to the server during launch
* `Server_MaxPlayers` - The maximum amount of players argument passed to the server during launch
* `Server_Game` - The gamemode alias argument passed to the server during launch
* `SNS_TOPIC` - The SNS topic to subscribe to and listen for updates on

When the server process is launched, this is how the tags are used

`Process.exe Server_Map?ServerName=Server_Name?MaxPlayers=Server_MaxPlayers?Game=Server_Game`
