class Checksum < Formula
  desc "SHA256 file verification for consistency check purposes"
  homepage "https://github.com/teran/checksum"
  url "https://github.com/teran/checksum/archive/v0.3.tar.gz"
  sha256 "fc5fc5f79b569067847fa420a19383a571dcc33b2e2c61d2413ed47afec1cc28"

  depends_on "go" => :build
  depends_on "make" => :build

  def install
    ENV["GOPATH"] = buildpath
    mkdir_p "src/github.com/teran"
    ln_s buildpath, "src/github.com/teran/checksum"
    system "make", "build-macos-amd64"
    system "mv", "bin/checksum-darwin-amd64", "bin/checksum"
    bin.install "bin/checksum"
  end

  test do
    system "make", "test"
  end
end
