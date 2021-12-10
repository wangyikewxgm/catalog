# KubeVela experimental addons

This dir is KubeVela official experimental addon-registry which will contain hasn't verified stable KubeVela addon.

Addon files in this dir files will be synced to alibaba-oss but will not be set in KubeVela by default.

You can add this addon registry by vela cli then use these addons.

```shell
$ vela addon registry add experimental --type=git --gitUrl=https://github.com/oam-dev/catalog/ --path=experimental/addons
```

```shell
$ vela addon registry list      
Name            Type    URL                                                                
KubeVela        Oss     https://addons.kubevela.net                                        
experimental    Git     https://github.com/oam-dev/catalog//tree/master/experimental/addons
```

