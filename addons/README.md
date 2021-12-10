# KubeVela addons

This dir is KubeVela official addon-registry which will contain stable KubeVela addon.

Addon files in this dir files will be synced to alibaba-oss which will be set in KubeVela as default addons registry.

You can list all addons by vela cli. For example

```shell
vela addon list
```

And you can manage the addon-registry by vela cli.

```shell
$ vela addon registry list 
Name            Type    URL                        
KubeVela        Oss     https://addons.kubevela.net

```

