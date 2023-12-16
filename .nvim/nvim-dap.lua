local dap = require("dap")

dap.adapters.delve = {
  type = "server",
  port = "${port}",
  executable = {
    command = "dlv",
    args = { "dap", "-l", "127.0.0.1:${port}" },
    options = {
      env = {
        GOPATH = "${workspaceFolder}/githooks/.go",
        GOBIN = "${workspaceFolder}/githooks/bin",
      },
    },
  },
}

dap.configurations.go = {
  -- This are the requests documented here:
  -- https://github.com/go-delve/delve/blob/master/Documentation/api/dap/README.md
  {
    type = "delve",
    name = "[githooks] Debug",
    request = "launch",
    program = "${file}",
  },

  -- works with go.mod packages and sub packages
  {
    type = "delve",
    name = "[githooks] Debug Test (go.mod)",
    request = "launch",
    mode = "test",
    program = "${fileDirname}",

    -- Because we are in a subdirevtory, this is needed.
    dlvCwd = "githooks",
  },
}
