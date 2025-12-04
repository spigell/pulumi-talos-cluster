# hcloud-ha-py

Example Pulumi program in Python that loads an HA cluster spec from `cluster.yaml` using the shared cluster loader and exports basic info. It mirrors the Go HA example.

## Running

```sh
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
pulumi up
```
