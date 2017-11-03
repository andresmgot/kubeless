# Usage

This image is used to serve as a proxy between the kubeless CLI and minio. It is required since it is not possible to access directly to Minio from outside the cluster.

# Build

Build this image executing the build script:
```
$ ./build.sh
```

# Run
Run this image specifying the namespace in which Minio is running specifying the "MINIO_NAMESPACE" env var.
