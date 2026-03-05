# Depozit Server

A secure gRPC file storage server.

## Installation

The easiest way to install is using the management script:

```bash
sudo ./depozit.sh
```

To uninstall:

```bash
sudo ./depozit.sh uninstall
sudo rm -rf /var/lib/depozit
```

**IMPORTANT**
The `uninstall` command **does not** remove your stored files. If you want to delete all data, you must manually remove the storage directory:

```bash
sudo rm -rf /var/lib/depozit
```
