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
# The path to steamcmd
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
