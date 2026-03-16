{
  "targets": [
    {
      "target_name": "kernel",
      "sources": ["kernel_napi.c"],
      "include_dirs": ["."],
      "conditions": [
        ["OS=='linux'", {
          "libraries": [
            "<(module_root_dir)/libkernel.a",
            "-lpthread", "-lm", "-ldl"
          ]
        }],
        ["OS=='mac'", {
          "libraries": [
            "<(module_root_dir)/libkernel.a",
            "-lpthread", "-lm", "-framework CoreFoundation",
            "-framework Security", "-framework SystemConfiguration"
          ]
        }],
        ["OS=='win'", {
          "libraries": [
            "<(module_root_dir)/libkernel.lib"
          ]
        }]
      ]
    }
  ]
}
