class Checksum < Formula
  desc "SHA256 file verification for consistency check purposes"
  homepage "https://github.com/teran/checksum"
  version "0.5"
  url "https://github.com/teran/checksum/archive/v#{version}.tar.gz"
  sha256 "aa5a7c50f549c4142ae56841c427edfbdd649259a7759fd05a243bf9f0bffb91"

  depends_on "go" => :build
  depends_on "make" => :build

  def install
    ENV["GOPATH"] = buildpath
    ENV["REVISION"] = version
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
