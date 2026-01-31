class SitiCli < Formula
  desc "个人命令行工具集"
  homepage "https://github.com/SeSiTing/homebrew-siti-cli"
  url "https://github.com/SeSiTing/homebrew-siti-cli/archive/v1.0.3.tar.gz"
  sha256 "2a109e830b9d7b80f8d858faf59c1723a768462ad70e06863e3de28cf075c5b3"
  license "MIT"

  def install
    bin.install "bin/siti"
    (share/"siti-cli").install "src/commands"
    (share/"siti-cli").install "scripts/post-install.sh"
    zsh_completion.install "completions/_siti" if File.exist?("completions/_siti")
    bash_completion.install "completions/siti.bash" if File.exist?("completions/siti.bash")
  end

  def post_install
    system "#{share}/siti-cli/scripts/post-install.sh"
  end

  test do
    assert_match "siti - 个人CLI工具集", shell_output("#{bin}/siti --help")
  end
end
