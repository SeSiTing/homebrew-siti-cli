class SitiCli < Formula
  desc "个人命令行工具集"
  homepage "https://github.com/SeSiTing/homebrew-siti-cli"
  url "https://github.com/SeSiTing/homebrew-siti-cli/archive/v1.2.20.tar.gz"
  sha256 "adf10b47535e77d77ec1933277e22f8c409950435382109724e1c76887add264"
  license "MIT"

  def install
    bin.install "bin/siti"
    (share/"siti-cli").install "src/commands"
    (share/"siti-cli/scripts").install "scripts/post-install.sh"
    (share/"siti-cli/scripts").install "scripts/post-uninstall.sh"
    (share/"siti-cli/scripts").install "scripts/migrate-to-unified.sh"
    zsh_completion.install "completions/_siti" if File.exist?("completions/_siti")
    bash_completion.install "completions/siti.bash" if File.exist?("completions/siti.bash")
  end

  def post_install
    system "#{share}/siti-cli/scripts/post-install.sh"
  end

  def caveats
    <<~EOS
      首次运行任意 siti 命令时会自动配置 shell wrapper。
      配置后重新打开终端或运行 source ~/.zshrc 即可生效。

      卸载后如需清理 ~/.zshrc 中的 siti 配置，请运行：
        siti cleanup

      运行 'siti --help' 查看所有命令
    EOS
  end

  test do
    assert_match "siti - 个人CLI工具集", shell_output("#{bin}/siti --help")
  end
end
