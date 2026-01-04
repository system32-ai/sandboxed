class Sandboxed < Formula
  desc "A comprehensive sandbox platform for secure code execution in Kubernetes environments"
  homepage "https://github.com/system32-ai/sandboxed"
  version "v1.0.8"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/system32-ai/sandboxed/releases/download/v1.0.8/sandboxed-v1.0.8-darwin-amd64.tar.gz"
      sha256 "0136dcc5316f2c9cf6209f005360e4feb868448f5bd226d82373aaf5b3155ac2"

      def install
        bin.install "sandboxed"
      end
    end
    if Hardware::CPU.arm?
      url "https://github.com/system32-ai/sandboxed/releases/download/v1.0.8/sandboxed-v1.0.8-darwin-arm64.tar.gz"
      sha256 "df8041d66f2f33bbcd2fe6daf3a46e56be762d19a685972faf55f44ffa7cc6d0"

      def install
        bin.install "sandboxed"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/system32-ai/sandboxed/releases/download/v1.0.8/sandboxed-v1.0.8-linux-amd64.tar.gz"
      # SHA256 will need to be updated manually for Linux builds
      # sha256 "LINUX_AMD64_SHA256_HERE"

      def install
        bin.install "sandboxed"
      end
    end
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/system32-ai/sandboxed/releases/download/v1.0.8/sandboxed-v1.0.8-linux-arm64.tar.gz"
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
