{
  description = "subs-check - A subscription checking tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        # 使用 nixpkgs 中最新的 Go 版本（支持 Go 1.25+ 特性）
        goVersion = pkgs.go;
      in
      {
        packages = {
          default = pkgs.buildGoModule {
            pname = "subs-check";
            version = self.shortRev or "dev";

            src = ./.;

            vendorHash = null; # go.sum 存在，使用 vendor 模式

            env.CGO_ENABLED = "0";

            ldflags = [
              "-s"
              "-w"
              "-X main.Version=${self.shortRev or "dev"}"
            ];

            meta = with pkgs.lib; {
              description = "Subscription checking tool";
              homepage = "https://github.com/sinspired/subs-check";
              license = licenses.agpl3Only;
              mainProgram = "subs-check";
            };
          };
        };

        devShells.default = pkgs.mkShell {
          name = "subs-check-dev";

          buildInputs = with pkgs; [
            # Go 工具链
            goVersion
            gopls
            gotools
            go-tools

            # 构建工具
            gnumake
            git

            # 交叉编译支持（可选）
            # gcc  # 如需 CGO 支持
          ];

          shellHook = ''
            echo "🚀 subs-check 开发环境"
            echo "Go 版本: $(go version)"
            echo ""
            echo "可用命令:"
            echo "  make build      - 构建当前平台二进制文件"
            echo "  make build-all  - 构建所有平台二进制文件"
            echo "  make clean      - 清理构建产物"
            echo "  make help       - 显示帮助信息"
            echo ""
            # 设置 GOPATH 到项目目录（可选）
            export GOPATH="$PWD/.gopath"
            export PATH="$GOPATH/bin:$PATH"
          '';

          # 保持与 Makefile 一致的环境变量
          CGO_ENABLED = "0";
        };
      });
}
