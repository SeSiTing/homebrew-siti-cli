class SitiCli < Formula
  desc "个人命令行工具集"
  homepage "https://github.com/SeSiTing/homebrew-siti-cli"
  url "https://github.com/SeSiTing/homebrew-siti-cli/archive/v1.0.9.tar.gz"
  sha256 "93b41236bee1b7899ee5185bcee0ca22f2ef7a32b6328c185d3b94ebcc53a1a0"
  license "MIT"

  def install
    bin.install "bin/siti"
    (share/"siti-cli").install "src/commands"
    (share/"siti-cli/scripts").install "scripts/post-install.sh"
    (share/"siti-cli/scripts").install "scripts/post-uninstall.sh"
    zsh_completion.install "completions/_siti" if File.exist?("completions/_siti")
    bash_completion.install "completions/siti.bash" if File.exist?("completions/siti.bash")
  end

  def post_install
    system "#{share}/siti-cli/scripts/post-install.sh"
  end

  def caveats
    <<~EOS
      ⚠️  重要：siti-cli 的部分命令需要配置 shell wrapper 才能在当前终端生效

      如果自动配置失败（权限问题），请手动运行：
        eval "$(siti init zsh)" >> ~/.zshrc
        source ~/.zshrc

      配置后，以下命令将在当前终端立即生效：
        • siti ai switch <provider>  - 切换 AI API 配置
        • siti proxy on/off          - 代理管理

      运行 'siti --help' 查看所有命令
    EOS
  end

  def post_uninstall
    system "#{share}/siti-cli/scripts/post-uninstall.sh"
  end

  test do
    assert_match "siti - 个人CLI工具集", shell_output("#{bin}/siti --help")
  end
end
