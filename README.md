# Hue Controller

Check if desired state and corrects it if necessary.

## Links

https://hue.quant.benjamin-borbe.de/lights

## Doc

https://developers.meethue.com/develop/get-started-2/

## Setup

Get bridge ip
https://discovery.meethue.com/

Create User/Token

```bash
curl \
--insecure \
-X POST \
-d '{"devicetype":"my_hue_app"}' \
https://<BRIDGE_IP>/api
```

Test the API

```bash
curl \
--insecure \
-X GET https://<BRIDGE_IP>/api/<YOUR_USERNAME>/lights
```

## Discover Clients

https://discovery.meethue.com/

```
[
    {
        "id": "ecb5fafffeab0260",
        "internalipaddress": "192.168.177.106",
        "port": 443
    },
    {
        "id": "ecb5fafffe1a4d5d",
        "internalipaddress": "192.168.177.118",
        "port": 443
    }
]
```

## License

BSD 2-Clause — see [LICENSE](LICENSE).
