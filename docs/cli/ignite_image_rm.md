## ignite image rm

Remove VM base images

### Synopsis


Remove one or multiple VM base images. Images are matched by prefix based on
their ID and name. To remove multiple images, chain the matches separated by spaces.
The force flag (-f, --force) kills and removes any running VMs using the image.


```
ignite image rm <image>... [flags]
```

### Options

```
  -f, --force   Force this operation. Warning, use of this mode may have unintended consequences.
  -h, --help    help for rm
```

### SEE ALSO

* [ignite image](ignite_image.md)	 - Manage VM base images

###### Auto generated by spf13/cobra on 7-Jun-2019