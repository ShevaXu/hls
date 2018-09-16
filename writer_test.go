package hls_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ShevaXu/hls"
)

func TestNewMediaPlaylist(t *testing.T) {
	_, e := hls.NewMediaPlaylist(1, 2)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	_, e = hls.NewMediaPlaylist(2, 1) //wrong winsize
	if e == nil {
		t.Fatal("Create new media playlist must be failed, but it's don't")
	}
}

// Create new media playlist
// Add two segments to media playlist
func TestAddSegmentToMediaPlaylist(t *testing.T) {
	p, e := hls.NewMediaPlaylist(1, 2)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	e = p.Append(hls.QuickSegment("test01.ts", "title", 10.0))
	if e != nil {
		t.Errorf("Add 1st segment to a media playlist failed: %s", e)
	}
	if p.Segments[0].URI != "test01.ts" {
		t.Errorf("Expected: test01.ts, got: %v", p.Segments[0].URI)
	}
	if p.Segments[0].Duration != 10 {
		t.Errorf("Expected: 10, got: %v", p.Segments[0].Duration)
	}
	if p.Segments[0].Title != "title" {
		t.Errorf("Expected: title, got: %v", p.Segments[0].Title)
	}
}

func TestAppendSegmentToMediaPlaylist(t *testing.T) {
	p, _ := hls.NewMediaPlaylist(2, 2)
	e := p.Append(&hls.MediaSegment{Duration: 10})
	if e != nil {
		t.Errorf("Add 1st segment to a media playlist failed: %s", e)
	}
	if p.TargetDuration != 10 {
		t.Errorf("Failed to increase TargetDuration, expected: 10, got: %v", p.TargetDuration)
	}
	e = p.Append(&hls.MediaSegment{Duration: 10})
	if e != nil {
		t.Errorf("Add 2nd segment to a media playlist failed: %s", e)
	}
	e = p.Append(&hls.MediaSegment{Duration: 10})
	if e != hls.ErrPlaylistFull {
		t.Errorf("Add 3rd expected full error, got: %s", e)
	}
}

// Create new media playlist
// Add three segments to media playlist
// Set discontinuity tag for the 2nd segment.
func TestDiscontinuityForMediaPlaylist(t *testing.T) {
	var e error
	p, e := hls.NewMediaPlaylist(3, 4)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	p.Close()
	if e = p.Append(hls.QuickSegment("test01.ts", "title", 5.0)); e != nil {
		t.Errorf("Add 1st segment to a media playlist failed: %s", e)
	}
	if e = p.Append(hls.QuickSegment("test02.ts", "title", 6.0)); e != nil {
		t.Errorf("Add 2nd segment to a media playlist failed: %s", e)
	}
	if e = p.SetDiscontinuity(); e != nil {
		t.Error("Can't set discontinuity tag")
	}
	if e = p.Append(hls.QuickSegment("test03.ts", "title", 6.0)); e != nil {
		t.Errorf("Add 3nd segment to a media playlist failed: %s", e)
	}
	//fmt.Println(p.Encode().String())
}

// Create new media playlist
// Add three segments to media playlist
// Set program date and time for 2nd segment.
// Set discontinuity tag for the 2nd segment.
func TestProgramDateTimeForMediaPlaylist(t *testing.T) {
	var e error
	p, e := hls.NewMediaPlaylist(3, 4)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	p.Close()
	if e = p.Append(hls.QuickSegment("test01.ts", "title", 5.0)); e != nil {
		t.Errorf("Add 1st segment to a media playlist failed: %s", e)
	}
	if e = p.Append(hls.QuickSegment("test02.ts", "title", 6.0)); e != nil {
		t.Errorf("Add 2nd segment to a media playlist failed: %s", e)
	}
	loc, _ := time.LoadLocation("Europe/Moscow")
	if e = p.SetProgramDateTime(time.Date(2010, time.November, 30, 16, 25, 0, 125*1e6, loc)); e != nil {
		t.Error("Can't set program date and time")
	}
	if e = p.SetDiscontinuity(); e != nil {
		t.Error("Can't set discontinuity tag")
	}
	if e = p.Append(hls.QuickSegment("test03.ts", "title", 6.0)); e != nil {
		t.Errorf("Add 3nd segment to a media playlist failed: %s", e)
	}
	//fmt.Println(p.Encode().String())
}

// Create new media playlist
// Add two segments to media playlist with duration 9.0 and 9.1.
// Target duration must be set to nearest greater integer (= 10).
func TestTargetDurationForMediaPlaylist(t *testing.T) {
	p, e := hls.NewMediaPlaylist(1, 2)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	e = p.Append(hls.QuickSegment("test01.ts", "title", 9.0))
	if e != nil {
		t.Errorf("Add 1st segment to a media playlist failed: %s", e)
	}
	e = p.Append(hls.QuickSegment("test02.ts", "title", 9.1))
	if e != nil {
		t.Errorf("Add 2nd segment to a media playlist failed: %s", e)
	}
	if p.TargetDuration < 10.0 {
		t.Errorf("Target duration must = 10 (nearest greater integer to durations 9.0 and 9.1)")
	}
}

// Create new media playlist with capacity 10 elements
// Try to add 11 segments to media playlist (oversize error)
func TestOverAddSegmentsToMediaPlaylist(t *testing.T) {
	p, e := hls.NewMediaPlaylist(1, 10)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	for i := 0; i < 11; i++ {
		e = p.Append(hls.QuickSegment(fmt.Sprintf("test%d.ts", i), "", 5.0))
		if e != nil {
			t.Logf("As expected new segment #%d not assigned to a media playlist: %s due oversize\n", i, e)
		}
	}
}

func TestSetSCTE35(t *testing.T) {
	p, _ := hls.NewMediaPlaylist(1, 2)
	scte := &hls.SCTE{Cue: "some cue"}
	if err := p.SetSCTE35(scte); err == nil {
		t.Error("SetSCTE35 expected empty playlist error")
	}
	_ = p.Append(hls.QuickSegment("test01.ts", "title", 10.0))
	if err := p.SetSCTE35(scte); err != nil {
		t.Errorf("SetSCTE35 did not expect error: %v", err)
	}
	if !reflect.DeepEqual(p.Segments[0].SCTE, scte) {
		t.Errorf("SetSCTE35\nexp: %#v\ngot: %#v", scte, p.Segments[0].SCTE)
	}
}

// Create new media playlist
// Add segment to media playlist
// Set SCTE
func TestSetSCTEForMediaPlaylist(t *testing.T) {
	tests := []struct {
		Cue      string
		ID       string
		Time     float64
		Expected string
	}{
		{"CueData1", "", 0, `#EXT-SCTE35:CUE="CueData1"` + "\n"},
		{"CueData2", "ID2", 0, `#EXT-SCTE35:CUE="CueData2",ID="ID2"` + "\n"},
		{"CueData3", "ID3", 3.141, `#EXT-SCTE35:CUE="CueData3",ID="ID3",TIME=3.141` + "\n"},
		{"CueData4", "", 3.1, `#EXT-SCTE35:CUE="CueData4",TIME=3.1` + "\n"},
		{"CueData5", "", 3.0, `#EXT-SCTE35:CUE="CueData5",TIME=3` + "\n"},
	}

	for _, test := range tests {
		p, e := hls.NewMediaPlaylist(1, 1)
		if e != nil {
			t.Fatalf("Create media playlist failed: %s", e)
		}
		if e = p.Append(hls.QuickSegment("test01.ts", "title", 5.0)); e != nil {
			t.Errorf("Add 1st segment to a media playlist failed: %s", e)
		}
		if e := p.SetSCTE35(&hls.SCTE{Syntax: hls.Syntax672014, Cue: test.Cue, ID: test.ID, Time: test.Time}); e != nil {
			t.Errorf("SetSCTE to a media playlist failed: %s", e)
		}
		if !strings.Contains(p.String(), test.Expected) {
			t.Errorf("Test %+v did not contain: %q, playlist: %v", test, test.Expected, p.String())
		}
	}
}

// Create new media playlist
// Add segment to media playlist
// Set encryption key
func TestSetKeyForMediaPlaylist(t *testing.T) {
	tests := []struct {
		KeyFormat         string
		KeyFormatVersions string
		ExpectVersion     int
	}{
		{"", "", 3},
		{"Format", "", 5},
		{"", "Version", 5},
		{"Format", "Version", 5},
	}

	for _, test := range tests {
		p, e := hls.NewMediaPlaylist(3, 5)
		if e != nil {
			t.Fatalf("Create media playlist failed: %s", e)
		}
		if e = p.Append(hls.QuickSegment("test01.ts", "title", 5.0)); e != nil {
			t.Errorf("Add 1st segment to a media playlist failed: %s", e)
		}
		if e := p.SetKey("AES-128", "https://example.com", "iv", test.KeyFormat, test.KeyFormatVersions); e != nil {
			t.Errorf("Set key to a media playlist failed: %s", e)
		}
		if p.Version() != test.ExpectVersion {
			t.Errorf("Set key playlist version: %v, expected: %v", p.Version(), test.ExpectVersion)
		}
	}
}

// Create new media playlist
// Add segment to media playlist
// Set encryption key
func TestSetDefaultKeyForMediaPlaylist(t *testing.T) {
	tests := []struct {
		KeyFormat         string
		KeyFormatVersions string
		ExpectVersion     int
	}{
		{"", "", 3},
		{"Format", "", 5},
		{"", "Version", 5},
		{"Format", "Version", 5},
	}

	for _, test := range tests {
		p, e := hls.NewMediaPlaylist(3, 5)
		if e != nil {
			t.Fatalf("Create media playlist failed: %s", e)
		}
		if e := p.SetDefaultKey("AES-128", "https://example.com", "iv", test.KeyFormat, test.KeyFormatVersions); e != nil {
			t.Errorf("Set key to a media playlist failed: %s", e)
		}
		if p.Version() != test.ExpectVersion {
			t.Errorf("Set key playlist version: %v, expected: %v", p.Version(), test.ExpectVersion)
		}
	}
}

// Create new media playlist
// Set default map
func TestSetDefaultMapForMediaPlaylist(t *testing.T) {
	p, e := hls.NewMediaPlaylist(3, 5)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	p.SetDefaultMap("https://example.com", 1000*1024, 1024*1024)

	expected := `EXT-X-MAP:URI="https://example.com",BYTERANGE=1024000@1048576`
	if !strings.Contains(p.String(), expected) {
		t.Fatalf("Media playlist did not contain: %s\nMedia Playlist:\n%v", expected, p.String())
	}
}

// Create new media playlist
// Add segment to media playlist
// Set map on segment
func TestSetMapForMediaPlaylist(t *testing.T) {
	p, e := hls.NewMediaPlaylist(3, 5)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	e = p.Append(hls.QuickSegment("test01.ts", "", 5.0))
	if e != nil {
		t.Errorf("Add 1st segment to a media playlist failed: %s", e)
	}
	e = p.SetMap("https://example.com", 1000*1024, 1024*1024)
	if e != nil {
		t.Errorf("Set map to a media playlist failed: %s", e)
	}

	expected := `EXT-X-MAP:URI="https://example.com",BYTERANGE=1024000@1048576
#EXTINF:5.000,
test01.ts`
	if !strings.Contains(p.String(), expected) {
		t.Fatalf("Media playlist did not contain: %s\nMedia Playlist:\n%v", expected, p.String())
	}
}

// Create new media playlist
// Set default map
// Add segment to media playlist
// Set map on segment (should be ignored when encoding)
func TestEncodeMediaPlaylistWithDefaultMap(t *testing.T) {
	p, e := hls.NewMediaPlaylist(3, 5)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	p.SetDefaultMap("https://example.com", 1000*1024, 1024*1024)

	e = p.Append(hls.QuickSegment("test01.ts", "title", 5.0))
	if e != nil {
		t.Errorf("Add 1st segment to a media playlist failed: %s", e)
	}
	e = p.SetMap("https://notencoded.com", 1000*1024, 1024*1024)
	if e != nil {
		t.Errorf("Set map to segment failed: %s", e)
	}

	encoded := p.String()
	expected := `EXT-X-MAP:URI="https://example.com",BYTERANGE=1024000@1048576`
	if !strings.Contains(encoded, expected) {
		t.Fatalf("Media playlist did not contain: %s\nMedia Playlist:\n%v", expected, encoded)
	}

	ignored := `EXT-X-MAP:URI="https://notencoded.com"`
	if strings.Contains(encoded, ignored) {
		t.Fatalf("Media playlist contains non default map: %s\nMedia Playlist:\n%v", ignored, encoded)
	}
}

// Create new media playlist
// Add two segments to media playlist
// Encode structures to HLS
func TestEncodeMediaPlaylist(t *testing.T) {
	p, e := hls.NewMediaPlaylist(3, 5)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	e = p.Append(hls.QuickSegment("test01.ts", "title", 5.0))
	if e != nil {
		t.Errorf("Add 1st segment to a media playlist failed: %s", e)
	}
	p.DurationAsInt(true)
	//fmt.Println(p.Encode().String())
}

// Create new media playlist
// Add 10 segments to media playlist
// Test iterating over segments
func TestLoopSegmentsOfMediaPlaylist(t *testing.T) {
	p, e := hls.NewMediaPlaylist(3, 5)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	for i := 0; i < 5; i++ {
		e = p.Append(hls.QuickSegment(fmt.Sprintf("test%d.ts", i), "", 5.0))
		if e != nil {
			t.Errorf("Add segment #%d to a media playlist failed: %s", i, e)
		}
	}
	p.DurationAsInt(true)
	//fmt.Println(p.Encode().String())
}

// Create new media playlist with capacity 5
// Add 5 segments and 5 unique keys
// Test correct keys set on correct segments
func TestEncryptionKeysInMediaPlaylist(t *testing.T) {
	p, _ := hls.NewMediaPlaylist(5, 5)
	// Add 5 segments and set custom encryption key
	for i := uint(0); i < 5; i++ {
		uri := fmt.Sprintf("uri-%d", i)
		expected := &hls.Key{
			Method:            "AES-128",
			URI:               uri,
			IV:                fmt.Sprintf("%d", i),
			Keyformat:         "identity",
			Keyformatversions: "1",
		}
		_ = p.Append(hls.QuickSegment(uri+".ts", "", 4))
		_ = p.SetKey(expected.Method, expected.URI, expected.IV, expected.Keyformat, expected.Keyformatversions)

		if p.Segments[i].Key == nil {
			t.Fatalf("Key was not set on segment %v", i)
		}
		if *p.Segments[i].Key != *expected {
			t.Errorf("Key %+v does not match expected %+v", p.Segments[i].Key, expected)
		}
	}
}

func TestEncryptionKeyMethodNoneInMediaPlaylist(t *testing.T) {
	p, e := hls.NewMediaPlaylist(5, 5)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	p.Append(hls.QuickSegment("segment-1.ts", "", 4))
	p.SetKey("AES-128", "key-uri", "iv", "identity", "1")
	p.Append(hls.QuickSegment("segment-2.ts", "", 4))
	p.SetKey("NONE", "", "", "", "")
	expected := `#EXT-X-KEY:METHOD=NONE
#EXTINF:4.000,
segment-2.ts`
	if !strings.Contains(p.String(), expected) {
		t.Errorf("Manifest %+v did not contain expected %+v", p, expected)
	}
}

// Create new media playlist
// Add 10 segments to media playlist
// Encode structure to HLS with integer target durations
func TestMediaPlaylistWithIntegerDurations(t *testing.T) {
	p, e := hls.NewMediaPlaylist(3, 10)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	for i := 0; i < 9; i++ {
		e = p.Append(hls.QuickSegment(fmt.Sprintf("test%d.ts", i), "", 5.6))
		if e != nil {
			t.Errorf("Add segment #%d to a media playlist failed: %s", i, e)
		}
	}
	p.DurationAsInt(false)
	//	fmt.Println(p.Encode().String())
}

// Create new media playlist
// Add 9 segments to media playlist
// 11 times encode structure to HLS with integer target durations
// Last playlist must be empty
func TestMediaPlaylistWithEmptyMedia(t *testing.T) {
	p, e := hls.NewMediaPlaylist(3, 10)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	for i := 1; i < 10; i++ {
		e = p.Append(hls.QuickSegment(fmt.Sprintf("test%d.ts", i), "", 5.6))
		if e != nil {
			t.Errorf("Add segment #%d to a media playlist failed: %s", i, e)
		}
	}
	for i := 1; i < 11; i++ {
		//fmt.Println(p.Encode().String())
		p.Remove()
	} // TODO add check for buffers equality
}

// Create new media playlist with winsize == capacity
func TestMediaPlaylistWinsize(t *testing.T) {
	p, e := hls.NewMediaPlaylist(6, 6)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	for i := 1; i < 10; i++ {
		p.Slide(hls.QuickSegment(fmt.Sprintf("test%d.ts", i), "", 5.6))
		//fmt.Println(p.Encode().String()) // TODO check playlist sizes and mediasequence values
	}
}

// Create new media playlist as sliding playlist.
// Close it.
func TestClosedMediaPlaylist(t *testing.T) {
	p, e := hls.NewMediaPlaylist(1, 10)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	for i := 0; i < 10; i++ {
		e = p.Append(hls.QuickSegment(fmt.Sprintf("test%d.ts", i), "", 5.0))
		if e != nil {
			t.Errorf("Due oversize new segment #%d not assigned to a media playlist: %s\n", i, e)
		}
	}
	p.Close()
}

// Create new media playlist as sliding playlist.
func TestLargeMediaPlaylistWithParallel(t *testing.T) {
	testCount := 10
	expect, err := ioutil.ReadFile("sample-playlists/media-playlist-large.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < testCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f, err := os.Open("sample-playlists/media-playlist-large.m3u8")
			if err != nil {
				t.Fatal(err)
			}
			p, err := hls.NewMediaPlaylist(50000, 50000)
			if err != nil {
				t.Fatalf("Create media playlist failed: %s", err)
			}
			if err = p.DecodeFrom(bufio.NewReader(f), true); err != nil {
				t.Fatal(err)
			}

			actual := p.Encode().Bytes() // disregard output
			if bytes.Compare(expect, actual) != 0 {
				t.Fatal("not matched")
			}
		}()
		wg.Wait()
	}
}

func TestMediaSetVersion(t *testing.T) {
	m, _ := hls.NewMediaPlaylist(3, 3)
	m.SetVersion(5)
	if m.Version() != 5 {
		t.Errorf("Expected version: %v, got: %v", 5, m.Version())
	}
}

func TestMediaWinSize(t *testing.T) {
	m, _ := hls.NewMediaPlaylist(3, 3)
	if m.WinSize() != 3 {
		t.Errorf("Expected winsize: %v, got: %v", 3, m.WinSize())
	}
}

func TestMediaSetWinSize(t *testing.T) {
	m, _ := hls.NewMediaPlaylist(3, 5)
	err := m.SetWinSize(5)
	if err != nil {
		t.Fatal(err)
	}
	if m.WinSize() != 5 {
		t.Errorf("Expected winsize: %v, got: %v", 5, m.WinSize())
	}
	// Check winsize cannot exceed capacity
	err = m.SetWinSize(99999)
	if err == nil {
		t.Error("Expected error, received: ", err)
	}
	// Ensure winsize didn't change
	if m.WinSize() != 5 {
		t.Errorf("Expected winsize: %v, got: %v", 5, m.WinSize())
	}
}

func TestMediaPlaylist_ExtendCapacity(t *testing.T) {
	m, _ := hls.NewMediaPlaylist(0, 2)
	err := m.ExtendCapacity()
	if err == nil {
		t.Errorf("Expected error, got:%v", err)
	}
	m.Append(hls.QuickSegment("1.ts", "", 3.1))
	m.Append(hls.QuickSegment("2.ts", "", 3.1))
	err = m.Append(hls.QuickSegment("3.ts", "", 3.1))
	if err == hls.ErrPlaylistFull {
		err = m.ExtendCapacity()
		if err != nil {
			t.Fatal(err)
		}
		m.Append(hls.QuickSegment("3.ts", "", 3.1))
	}
	if m.Count() != 3 {
		t.Errorf("Expected count:%v, got:%v", 3, m.Count())
	}
}

func TestMediaPlaylist_AppendWithAutoExtend(t *testing.T) {
	m, _ := hls.NewMediaPlaylist(0, 2)
	m.Append(hls.QuickSegment("1.ts", "", 3.1))
	m.Append(hls.QuickSegment("2.ts", "", 3.1))
	err := m.AppendWithAutoExtend(hls.QuickSegment("3.ts", "", 3.1))
	if err != nil {
		t.Fatal(err)
	}
	if m.Count() != 3 {
		t.Errorf("Expected count:%v, got:%v", 3, m.Count())
	}
}

// Create new master playlist without params
// Add media playlist
func TestNewMasterPlaylist(t *testing.T) {
	m := hls.NewMasterPlaylist()
	p, e := hls.NewMediaPlaylist(3, 5)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	for i := 0; i < 5; i++ {
		e = p.Append(hls.QuickSegment(fmt.Sprintf("test%d.ts", i), "", 5.0))
		if e != nil {
			t.Errorf("Add segment #%d to a media playlist failed: %s", i, e)
		}
	}
	m.Append("chunklist1.m3u8", p, hls.VariantParams{})
}

// Create new master playlist without params
// Add media playlist with Alternatives
func TestNewMasterPlaylistWithAlternatives(t *testing.T) {
	m := hls.NewMasterPlaylist()
	audioUri := fmt.Sprintf("%s/rendition.m3u8", "800")
	audioAlt := &hls.Alternative{
		GroupID:    "audio",
		URI:        audioUri,
		Type:       "AUDIO",
		Name:       "main",
		Default:    true,
		Autoselect: "YES",
		Language:   "english",
	}
	p, e := hls.NewMediaPlaylist(3, 5)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	for i := 0; i < 5; i++ {
		e = p.Append(hls.QuickSegment(fmt.Sprintf("test%d.ts", i), "", 5.0))
		if e != nil {
			t.Errorf("Add segment #%d to a media playlist failed: %s", i, e)
		}
	}
	m.Append("chunklist1.m3u8", p, hls.VariantParams{Alternatives: []*hls.Alternative{audioAlt}})

	if m.Version() != 4 {
		t.Fatalf("Expected version 4, actual, %d", m.Version())
	}
	expected := `#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",NAME="main",DEFAULT=YES,AUTOSELECT=YES,LANGUAGE="english",URI="800/rendition.m3u8"`
	if !strings.Contains(m.String(), expected) {
		t.Fatalf("Master playlist did not contain: %s\nMaster Playlist:\n%v", expected, m.String())
	}
}

// Create new master playlist supporting CLOSED-CAPTIONS=NONE
func TestNewMasterPlaylistWithClosedCaptionEqNone(t *testing.T) {
	m := hls.NewMasterPlaylist()

	vp := &hls.VariantParams{
		ProgramID:  0,
		Bandwidth:  8000,
		Codecs:     "avc1",
		Resolution: "1280x720",
		Audio:      "audio0",
		Captions:   "NONE",
	}

	p, err := hls.NewMediaPlaylist(1, 1)
	if err != nil {
		t.Fatalf("Create media playlist failed: %s", err)
	}
	m.Append(fmt.Sprintf("eng_rendition_rendition.m3u8"), p, *vp)

	expected := "CLOSED-CAPTIONS=NONE"
	if !strings.Contains(m.String(), expected) {
		t.Fatalf("Master playlist did not contain: %s\nMaster Playlist:\n%v", expected, m.String())
	}
	// quotes need to be include if not eq NONE
	vp.Captions = "CC1"
	m2 := hls.NewMasterPlaylist()
	m2.Append(fmt.Sprintf("eng_rendition_rendition.m3u8"), p, *vp)
	expected = `CLOSED-CAPTIONS="CC1"`
	if !strings.Contains(m2.String(), expected) {
		t.Fatalf("Master playlist did not contain: %s\nMaster Playlist:\n%v", expected, m2.String())
	}
}

// Create new master playlist with params
// Add media playlist
func TestNewMasterPlaylistWithParams(t *testing.T) {
	m := hls.NewMasterPlaylist()
	p, e := hls.NewMediaPlaylist(3, 5)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	for i := 0; i < 5; i++ {
		e = p.Append(hls.QuickSegment(fmt.Sprintf("test%d.ts", i), "", 5.0))
		if e != nil {
			t.Errorf("Add segment #%d to a media playlist failed: %s", i, e)
		}
	}
	m.Append("chunklist1.m3u8", p, hls.VariantParams{ProgramID: 123, Bandwidth: 1500000, Resolution: "576x480"})
}

// Create new master playlist
// Add media playlist with existing query params in URI
// Append more query params and ensure it encodes correctly
func TestEncodeMasterPlaylistWithExistingQuery(t *testing.T) {
	m := hls.NewMasterPlaylist()
	p, e := hls.NewMediaPlaylist(3, 5)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	for i := 0; i < 5; i++ {
		e = p.Append(hls.QuickSegment(fmt.Sprintf("test%d.ts", i), "", 5.0))
		if e != nil {
			t.Errorf("Add segment #%d to a media playlist failed: %s", i, e)
		}
	}
	m.Append("chunklist1.m3u8?k1=v1&k2=v2", p, hls.VariantParams{ProgramID: 123, Bandwidth: 1500000, Resolution: "576x480"})
	m.Args = "k3=v3"
	if !strings.Contains(m.String(), `chunklist1.m3u8?k1=v1&k2=v2&k3=v3`) {
		t.Errorf("Encode master with existing args failed")
	}
}

// Create new master playlist
// Add media playlist
// Encode structures to HLS
func TestEncodeMasterPlaylist(t *testing.T) {
	m := hls.NewMasterPlaylist()
	p, e := hls.NewMediaPlaylist(3, 5)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	for i := 0; i < 5; i++ {
		e = p.Append(hls.QuickSegment(fmt.Sprintf("test%d.ts", i), "", 5.0))
		if e != nil {
			t.Errorf("Add segment #%d to a media playlist failed: %s", i, e)
		}
	}
	m.Append("chunklist1.m3u8", p, hls.VariantParams{ProgramID: 123, Bandwidth: 1500000, Resolution: "576x480"})
	m.Append("chunklist2.m3u8", p, hls.VariantParams{ProgramID: 123, Bandwidth: 1500000, Resolution: "576x480"})
}

// Create new master playlist with Name tag in EXT-X-STREAM-INF
func TestEncodeMasterPlaylistWithStreamInfName(t *testing.T) {
	m := hls.NewMasterPlaylist()
	p, e := hls.NewMediaPlaylist(3, 5)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
	for i := 0; i < 5; i++ {
		e = p.Append(hls.QuickSegment(fmt.Sprintf("test%d.ts", i), "", 5.0))
		if e != nil {
			t.Errorf("Add segment #%d to a media playlist failed: %s", i, e)
		}
	}
	m.Append("chunklist1.m3u8", p, hls.VariantParams{ProgramID: 123, Bandwidth: 3000000, Resolution: "1152x960", Name: "HD 960p"})

	if m.Variants[0].Name != "HD 960p" {
		t.Fatalf("Create master with Name in EXT-X-STREAM-INF failed")
	}
	if !strings.Contains(m.String(), `NAME="HD 960p"`) {
		t.Fatalf("Encode master with Name in EXT-X-STREAM-INF failed")
	}
}

func TestMasterSetVersion(t *testing.T) {
	m := hls.NewMasterPlaylist()
	m.SetVersion(5)
	if m.Version() != 5 {
		t.Errorf("Expected version: %v, got: %v", 5, m.Version())
	}
}
