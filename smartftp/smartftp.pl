#!/usr/bin/perl

# Enhanced FTP send/receive files script
# Copyright (c) 2013 Nicola Ruggero <nicola@nxnt.org>
#
# This scripts send and get a pattern of files handling proper file rename to
# permits complete tranfers (avoid application to read files before tranfer ends)
#
# Usage: smartftp.pl --action [put|get] --host <ftphost> [--port <ftpport> ] 
#                    [ --username <username> --password <password> ]
#                    [ --rename_prefix <rename_prefix> --rename_suffix <rename_suffix>]
#                    --file <file1> --file <file2>... --remote_dir <remote_dir>
#
# ====================================================================
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.
# ====================================================================

use strict;
use warnings;
use Net::FTP;
use Getopt::Long;
use File::Basename;

our $VERSION = "1.0.0";


###############################################################################
# Initialize application

my $action;
my $ftp_host;
my $ftp_port = "21";
my $username = "anonymous";
my $password = 'aaa@bbb.com';
my $rename_prefix = "";
my $rename_suffix = "";
my @files;
my $remote_dir = ".";
my $ftp;


# Parse command line
if (@ARGV > 0 )
  {
    GetOptions(	'action=s'  		=> \$action,
				'host=s' 			=> \$ftp_host,
				'port=i'     		=> \$ftp_port,
				'username=s'   		=> \$username,
				'password=s'  		=> \$password,
				'rename_prefix=s'   => \$rename_prefix,
				'rename_suffix=s'   => \$rename_suffix,
				'file=s'			=> \@files,
				'remote_dir=s'     	=> \$remote_dir)
	     or usage();
  }
else 
  {
    usage();
  }

$username = ($username) ? $username : 'anonymous';
$password = ($password) ? $password : 'aaa@bbb.com';
$ftp_port = ($ftp_port) ? $ftp_port : '21';
usage() if ($action !~ /^get$|^put$/);

print scalar(localtime()) . " Starting smartftp $VERSION\n";


###############################################################################
# Main application

# Connect to remote server and start data transfer

print scalar(localtime()) . " Connecting to $ftp_host... ";
$ftp = Net::FTP->new($ftp_host, ('Port' => $ftp_port, 'Debug' => 0, 'Timeout' => 600))
	or die "KO\nCannot connect: $@";
print "OK\n";

print scalar(localtime()) . " Loggin in as $username... ";
$ftp->login($username,$password)
	or die "KO\nCannot login ", $ftp->message;
print "OK\n";

print scalar(localtime()) . " Changing remote directory $remote_dir... ";
$ftp->cwd($remote_dir)
	or die "KO\nCannot change remote directory ", $ftp->message;
print "OK\n";

print "\n";

foreach my $file (@files) {

	if ($action eq 'get')
		{
			# Actual file transfer
			print scalar(localtime()) . " GET $file\n";
			$ftp->get($file, $rename_prefix . $file . $rename_suffix)
				or die scalar(localtime()) . "  Get failed " . $ftp->message;
			# File rename after complete transfer
			rename $rename_prefix . $file . $rename_suffix, $file
				or die scalar(localtime()) . "  Rename failed " . $@ ."\n";
		}
	if ($action eq 'put')
		{
			print scalar(localtime()) . " PUT $file\n";
			$ftp->put($file, $rename_prefix . basename($file) . $rename_suffix)
				or die scalar(localtime()) . "  Put failed " . $ftp->message;
			# File rename after complete transfer
			$ftp->rename($rename_prefix . basename($file) . $rename_suffix, basename($file))
				or die scalar(localtime()) . "  Rename failed " . $ftp->message . $@;
		}
}

print scalar(localtime()) . " Disconnecting from $ftp_host... ";
$ftp->quit;
print "OK\n\n";

print scalar(localtime()) . " Smartftp completed :-)\n";
exit 0;


###############################################################################
# Functions

sub usage {
    die <<'EOF'
Smartftp: Enhanced FTP send/receive files script

Usage: smartftp.pl --action [put|get] --host <ftphost> [--port <ftpport> ] 
                    [ --username <username> --password <password> ]
                    [ --rename_prefix <rename_prefix> --rename_suffix <rename_suffix>]
                    --file <file1> --file <file2>... --remote_dir <remote_dir>
EOF
}