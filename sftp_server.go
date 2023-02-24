package sftp_server

import (
	"os"
	"time"
	"bytes"
	"strings"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)


type SFTPClient struct {
	Username string
	Password string
	IPAddress string
	Port string
}

type fileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
	Sys     interface{}
}

func (c *SFTPClient) connect() (*sftp.Client, error) {
	// Set up SSH configuration
	config := &ssh.ClientConfig{
		User: c.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to the SFTP server
	conn, err := ssh.Dial("tcp", c.IPAddress + ":" + c.Port, config)
	if err != nil {
		return nil, err
	}

	// Open an SFTP client session
	client, err := sftp.NewClient(conn)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *SFTPClient) AppendToFile(filePath string, data string) error {
	client, err := c.connect()
	if err != nil {
		return err
	}
	defer client.Close()

	// Check if the file exists
	_, err = client.Stat(filePath)
	if err == nil {
		// File exists, append to it
		f, err := client.OpenFile(filePath, os.O_APPEND|os.O_WRONLY)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = f.Write([]byte(data))
		if err != nil {
			return err
		}
		return nil
	}

	// File does not exist, create it
	f, err := client.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(data))
	if err != nil {
		return err
	}

	return nil
}

func (c *SFTPClient) OverwriteFile(filePath string, data string) error {
	client, err := c.connect()
	if err != nil {
		return err
	}
	defer client.Close()

	// Overwrite the file
	f, err := client.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(data))
	if err != nil {
		return err
	}

	return nil
}

func (c *SFTPClient) ReadFile(filePath string) ([]byte, error) {
	client, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// Open the file for reading
	f, err := client.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read all the lines in the file
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(f)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *SFTPClient) ListOfFilesDir(dirPath string) ([]os.FileInfo, error) {
	client, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// List the files and directories in the specified directory
	files,	err := client.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (c *SFTPClient) ListAllFiles(dirPath string) ([]fileInfo, error) {
	client, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// Recursively list all files and directories in the specified directory
	var allFiles []fileInfo
	err = c.listAllFilesRecursive(dirPath, "", client, &allFiles)
	if err != nil {
		return nil, err
	}

	return allFiles, nil
}

func (c *SFTPClient) listAllFilesRecursive(dirPath string, prefix string, client *sftp.Client, allFiles *[]fileInfo) error {
    files, err := client.ReadDir(dirPath)
    if err != nil {
        return err
    }
    for _, f := range files {
        if f.IsDir() {
            newPrefix := prefix + "/" + f.Name()
            err := c.listAllFilesRecursive(dirPath + "/" + f.Name(), newPrefix, client, allFiles)
            if err != nil {
                return err
            }
        } else {
			// Create a new FileInfo struct with the updated Name field
			newFile := &fileInfo{
				name:    prefix + "/" + f.Name(),
				size:    f.Size(),
				mode:    f.Mode(),
				modTime: f.ModTime(),
				isDir:   f.IsDir(),
				Sys: f.Sys(),
			}
			// Add the new FileInfo to the allFiles slice
			*allFiles = append(*allFiles, *newFile)
        }
    }

    return nil
}

func (c *SFTPClient) CreateDirectoryIfNotExist(dirPath string) error {
	client, err := c.connect()
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.Stat(dirPath)
	if err == nil {
		// Directory already exists, nothing to do
		return nil
	}

	// Directory does not exist, create it
	err = client.Mkdir(dirPath)
	if err != nil {
		return err
	}

	return nil
}

func (c *SFTPClient) CreateDirectoryRecursively(dirPath string) error {
	client, err := c.connect()
	if err != nil {
		return err
	}
	defer client.Close()

	// Split the directory path into individual components
	pathComponents := strings.Split(dirPath, "/")

	// Iterate through each path component and create the directories as needed
	currentPath := ""
	for _, component := range pathComponents {
		if component == "" {
			// Skip empty path components (e.g. from leading/trailing slashes)
			continue
		}
		currentPath += "/" + component
		_, err := client.Stat(currentPath)
		if err == nil {
			// Directory already exists, nothing to do
			continue
		}
		err = client.Mkdir(currentPath)
		if err != nil {
			return err
		}
	}
	return nil
}
