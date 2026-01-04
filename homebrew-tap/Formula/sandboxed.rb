class Sandboxed < Formula
  desc "A comprehensive sandbox platform for secure code execution in Kubernetes environments"
  homepage "https://github.com/system32-ai/sandboxed"
  version "v1.0.7"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/system32-ai/sandboxed/releases/download/v1.0.7/sandboxed-v1.0.7-darwin-amd64.tar.gz"
      sha256 "de86e5889c1b111dc33084d9323e7123719e14df45a3bdb97db15ba73b451d7e"

      def install
        bin.install "sandboxed"
      end
    end
    if Hardware::CPU.arm?
      url "https://github.com/system32-ai/sandboxed/releases/download/v1.0.7/sandboxed-v1.0.7-darwin-arm64.tar.gz"
      sha256 "cd2ad35b0d8f98d5cc5881eea5f289f27b3d0fa7a041768c69861781203302f7"

      def install
        bin.install "sandboxed"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/system32-ai/sandboxed/releases/download/v1.0.7/sandboxed-v1.0.7-linux-amd64.tar.gz"
      # SHA256 will need to be updated manually for Linux builds
      # sha256 "LINUX_AMD64_SHA256_HERE"

      def install
        bin.install "sandboxed"
      end
    end
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/system32-ai/sandboxed/releases/download/v1.0.7/sandboxed-v1.0.7-linux-arm64.tar.gz"
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
