class SitiCli < Formula
  desc "个人命令行工具集"
  homepage "https://github.com/SeSiTing/homebrew-siti-cli"
  url "https://github.com/SeSiTing/homebrew-siti-cli/archive/v1.0.7.tar.gz"
  sha256 "52c2030e88d047448b4a7bc51d0a42abb9e184ad233a0300857c9f4c662c7e97"
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

  def post_uninstall
    system "#{share}/siti-cli/scripts/post-uninstall.sh"
  end

  test do
    assert_match "siti - 个人CLI工具集", shell_output("#{bin}/siti --help")
  end
end
