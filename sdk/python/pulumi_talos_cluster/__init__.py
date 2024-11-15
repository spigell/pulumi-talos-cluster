# coding=utf-8
# *** WARNING: this file was generated by Pulumi SDK Generator. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

from . import _utilities
import typing
# Export this package's modules as members:
from ._enums import *
from .apply import *
from .cluster import *
from .provider import *
from ._inputs import *
from . import outputs
_utilities.register(
    resource_modules="""
[
 {
  "pkg": "talos-cluster",
  "mod": "index",
  "fqn": "pulumi_talos_cluster",
  "classes": {
   "talos-cluster:index:Apply": "Apply",
   "talos-cluster:index:Cluster": "Cluster"
  }
 }
]
""",
    resource_packages="""
[
 {
  "pkg": "talos-cluster",
  "token": "pulumi:providers:talos-cluster",
  "fqn": "pulumi_talos_cluster",
  "class": "Provider"
 }
]
"""
)
