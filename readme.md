# TwitterFollowersBot

TwitterFollowersBot is a little bot fot Twitter to check your new follows and unfollows. The bot will check your list of followers every "refreshtime" minutes and will send a DM telling you how started/stopped following you.

You must configure your bot by using the config.json file, and then run the program by running:

`bin/TwitterFollowersBot -c path/to/config.json`

### Installation

1. Clone this repository and rename the folder to TwitterFollowersBot if it is not the name.
2. Run the install script `install.sh` to install all the dependencies.
3. Compile using Makefile (`make`).

Finally, run using the executable for your platform.

**Tested on Go 1.7.3**
