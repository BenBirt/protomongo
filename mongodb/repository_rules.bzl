def _mongodb(ctx):
    os_dir_name = "linux"
    os_build_name = "linux-x86_64-debian10"
    if ctx.os.name.startswith("mac"):
        os_dir_name = "osx"
        os_build_name = "macos-x86_64"
    tar_file = "mongodb-%s-%s" % (os_build_name, ctx.attr.version)
    ctx.download_and_extract(
        "https://fastdl.mongodb.org/%s/%s.tgz" % (os_dir_name, tar_file),
        stripPrefix = tar_file,
    )
    ctx.file("BUILD", 'exports_files(glob(["bin/*"]))')

mongodb = repository_rule(
    implementation = _mongodb,
    attrs = {
        "version": attr.string(default = "4.2.2"),
    },
)
