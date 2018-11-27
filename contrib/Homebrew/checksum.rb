class Checksum < Formula
  desc "SHA256 file verification for consistency check purposes"
  homepage "https://github.com/teran/checksum"
  version "0.8.1"
  url "https://github.com/teran/checksum/archive/v#{version}.tar.gz"
  sha256 "fb2b3d7cc48328af82ae465ebd4daf64f99889f86711b68b834ef59e6a48db1c"

  depends_on "go" => :build
  depends_on "make" => :build
  depends_on "dep" => :build

  def install
    ENV["GOPATH"] = buildpath
    ENV["REVISION"] = version
    arch = MacOS.prefer_64_bit? ? "amd64" : "i386"
    (buildpath/"src/github.com/teran/checksum").install buildpath.children
    cd "src/github.com/teran/checksum" do
      ENV["DEP_BUILD_PLATFORMS"] = "darwin"
      ENV["DEP_BUILD_ARCHS"] = arch
      system "dep", "ensure"
      system "make", "build-macos-#{arch}"
      bin.install "bin/checksum-darwin-#{arch}" => "checksum"
      prefix.install_metafiles
    end
  end

  test do
    system "make", "test"
  end
end
