# govirtfsstat #

govirtfsstat allows to interact with files created by qemu when using 9p or virtiofs file share.

In order to modify file permissions and file types, data is saved in the extended attributes.
For more info see: https://github.com/qemu/qemu/blob/master/hw/9pfs/9p-local.c#L185
