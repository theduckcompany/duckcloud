# Systemd configuration

This folder contains an example of unit file that can be used both by a user or by a package manager.

The `duckcloud.service` file contains all the generic configurations, like the sandboxing rules and all the necessary 
restrictions to run Duckcloud securely. Its content is susceptible to change with each version so the package managers
should package this file with the binary at each versions and a user should update this file after each update.

This unit file assumes that:
- A user name `duckcloud` have been created.
- A file `/etc/duckcloud/var_file` exists with the `DATADIR` env variable setup. You can find [an example file here](./var_file.example).
- The `DATADIR` variable content is a path pointing to a folder owned by the `duckcloud` user.
