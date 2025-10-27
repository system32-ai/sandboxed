class Sandboxed < Formula
  desc "A comprehensive sandbox platform for secure code execution in Kubernetes environments"
  homepage "https://github.com/system32-ai/sandboxed"
  version "v1.0.6"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/system32-ai/sandboxed/releases/download/v1.0.6/sandboxed-v1.0.6-darwin-amd64.tar.gz"
      sha256 "d95910b068e28a9a610017fa6b47523eade395264b4df3ae281245a49f50f619"

      def install
        bin.install "sandboxed"
      end
    end
    if Hardware::CPU.arm?
      url "https://github.com/system32-ai/sandboxed/releases/download/v1.0.6/sandboxed-v1.0.6-darwin-arm64.tar.gz"
      sha256 "588e6fcaf7e8726d8b557ae49cb217c324cf97fdc9be1ff1307623b7a3df7521"

      def install
        bin.install "sandboxed"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/system32-ai/sandboxed/releases/download/v1.0.6/sandboxed-v1.0.6-linux-amd64.tar.gz"
      # SHA256 will need to be updated manually for Linux builds
      # sha256 "LINUX_AMD64_SHA256_HERE"

      def install
        bin.install "sandboxed"
      end
    end
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/system32-ai/sandboxed/releases/download/v1.0.6/sandboxed-v1.0.6-linux-arm64.tar.gz"
      # SHA256 will need to be updated manually for Linux builds
      # sha256 "LINUX_ARM64_SHA256_HERE"

      def install
        bin.install "sandboxed"
      end
    end
  end

  def caveats
    <<~EOS
      Sandboxed requires access to a Kubernetes cluster to function properly.
      
      Make sure you have kubectl configured and the necessary RBAC permissions:
      - pods: create, delete, get, list, watch
      - pods/exec: create
      
      For more information, visit: https://github.com/system32-ai/sandboxed
    EOS
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/sandboxed version")
  end
end
