class Claudectl < Formula
  include Language::Python::Virtualenv

  desc "CLI tool for managing Claude Code configurations and workspaces"
  homepage "https://github.com/carelesslisper/claudectl"
  url "https://github.com/carelesslisper/claudectl/releases/download/v0.1.0/claudectl-0.1.0.tar.gz"
  sha256 "PLACEHOLDER_UPDATE_AFTER_FIRST_RELEASE"
  license "MIT"

  depends_on "python@3.13"

  resource "docker" do
    url "https://files.pythonhosted.org/packages/source/d/docker/docker-7.1.0.tar.gz"
    sha256 "896c4282e5c7af5c45e8b683b0b3c33b481c5d1ca4c5b5b86d4de7924c42d7fa"
  end

  resource "gitpython" do
    url "https://files.pythonhosted.org/packages/source/g/gitpython/GitPython-3.1.45.tar.gz"
    sha256 "fce760879cd2aebd2991b3542876dc5c4a909b30c9d69dfc488e504a8db37ee8"
  end

  resource "typer" do
    url "https://files.pythonhosted.org/packages/source/t/typer/typer-0.20.0.tar.gz"
    sha256 "5b61a19d150d5a07e119a5b2c7c9b0e0f5db1fdd4e98e5aa21a05bb25308bbd3"
  end

  def install
    virtualenv_install_with_resources
  end

  test do
    assert_match "claudectl", shell_output("#{bin}/claudectl --help")
    assert_match "0.1.0", shell_output("#{bin}/claudectl version")
  end
end
