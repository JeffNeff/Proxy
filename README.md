# Vault Proxy

## Overview

vault proxy ssh's into ff3 and then forwards any traffic it sees on route / to port 8200 on the vault server.

export VAULT_SERVER=34.125.87.222:22
export KEY_PATH=/Users/jeffreynaef/.ssh/id_rsa