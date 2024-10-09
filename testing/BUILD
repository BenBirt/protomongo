load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["mongodb.go"],
    data = ["@mongodb//:bin/mongod"],
    importpath = "github.com/BenBirt/protomongo/testing",
    visibility = ["//visibility:public"],
    deps = [
        "//:go_default_library",
        "@com_github_golang_protobuf//proto:go_default_library",
        "@com_github_google_uuid//:go_default_library",
        "@com_github_phayes_freeport//:go_default_library",
        "@org_mongodb_go_mongo_driver//bson:go_default_library",
        "@org_mongodb_go_mongo_driver//mongo:go_default_library",
        "@org_mongodb_go_mongo_driver//mongo/options:go_default_library",
    ],
)
