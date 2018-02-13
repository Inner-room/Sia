package renter

import (
	"testing"

	"github.com/NebulousLabs/Sia/siatest"
)

// TestRenter executes a number of subtests using the same TestGroup to
// save time on initialization
func TestRenter(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// Create a group for the subtests
	groupParams := siatest.GroupParams{
		Hosts:   5,
		Renters: 1,
		Miners:  1,
	}
	tg, err := siatest.NewGroupFromTemplate(groupParams)
	if err != nil {
		t.Fatal("Failed to create group: ", err)
	}
	defer func() {
		if err := tg.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// Specifiy subtests to run
	subTests := []struct {
		name string
		test func(*testing.T, *siatest.TestGroup)
	}{
		{"UploadDownload", testUploadDownload},
	}
	// Run subtests
	for _, subtest := range subTests {
		t.Run(subtest.name, func(t *testing.T) {
			subtest.test(t, tg)
		})
	}
}

// testUploadDownload is a subtest that uses an existing TestGroup to test if
// uploading and downloading a file works
func testUploadDownload(t *testing.T, tg *siatest.TestGroup) {
	// Grab the first of the group's renters
	renter := tg.Renters()[0]
	// Upload file, creating a piece for each host in the group
	dataPieces := uint64(1)
	parityPieces := uint64(len(tg.Hosts())) - dataPieces
	file, err := renter.UploadNewFileBlocking(100, dataPieces, parityPieces)
	if err != nil {
		t.Fatal("Failed to create file for testing: ", err)
	}
	// Download the file synchronously directly into memory and compare the
	// data to the original
	_, err = renter.DownloadByStream(file)
	if err != nil {
		t.Fatal(err)
	}
	// Download the file synchronously to a file on disk and compare it to the
	// original
	_, err = renter.DownloadToDisk(file, false)
	if err != nil {
		t.Fatal(err)
	}
	// Download the file  asynchronously, wait for the download to finish and
	// compare it to the original
	downloadingFile, err := renter.DownloadToDisk(file, true)
	if err != nil {
		t.Error(err)
	}
	if err := renter.WaitForDownload(downloadingFile, file); err != nil {
		t.Error(err)
	}
}
