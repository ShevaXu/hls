package hls

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// ErrPlaylistFull is the error surfaced if a playlist is full.
// Consider extending the capacity in that situation.
var ErrPlaylistFull = errors.New("playlist is full")

// checkVersion checks and sets the playlist version accordingly with section 7.
func checkVersion(ver *int, newver int) {
	if *ver < newver {
		*ver = newver
	}
}

// QuickSegment returns a segment ready to append.
func QuickSegment(uri, title string, duration float64) *MediaSegment {
	return &MediaSegment{
		URI:      uri,
		Title:    title,
		Duration: duration,
	}
}

// NewMasterPlaylist creates a new empty master playlist.
func NewMasterPlaylist() *MasterPlaylist {
	p := new(MasterPlaylist)
	p.ver = minVer
	return p
}

// Append appends variant to master playlist.
// This operation does reset playlist cache.
func (p *MasterPlaylist) Append(uri string, chunklist *MediaPlaylist, params VariantParams) {
	v := new(Variant)
	v.URI = uri
	v.Chunklist = chunklist
	v.VariantParams = params
	p.Variants = append(p.Variants, v)
	if len(v.Alternatives) > 0 {
		// From section 7:
		// The EXT-X-MEDIA tag and the AUDIO, VIDEO and SUBTITLES attributes of
		// the EXT-X-STREAM-INF tag are backward compatible to protocol version
		// 1, but playback on older clients may not be desirable.  A server MAY
		// consider indicating a EXT-X-VERSION of 4 or higher in the Master
		// Playlist but is not required to do so.
		checkVersion(&p.ver, 4) // so it is optional and in theory may be set to ver.1
		// but more tests required
	}
	p.buf.Reset()
}

// ResetCache resets the underlying bytes buffer.
func (p *MasterPlaylist) ResetCache() {
	p.buf.Reset()
}

// Encode generates the output in M3U8 format.
func (p *MasterPlaylist) Encode() *bytes.Buffer {
	if p.buf.Len() > 0 {
		return &p.buf
	}

	p.buf.WriteString("#EXTM3U\n#EXT-X-VERSION:")
	p.buf.WriteString(strconv.Itoa(p.ver))
	p.buf.WriteRune('\n')

	var altsWritten = make(map[string]bool)

	for _, pl := range p.Variants {
		if pl.Alternatives != nil {
			for _, alt := range pl.Alternatives {
				// Make sure that we only write out an alternative once
				altKey := fmt.Sprintf("%s-%s-%s-%s", alt.Type, alt.GroupID, alt.Name, alt.Language)
				if altsWritten[altKey] {
					continue
				}
				altsWritten[altKey] = true

				p.buf.WriteString("#EXT-X-MEDIA:")
				if alt.Type != "" {
					p.buf.WriteString("TYPE=") // Type should not be quoted
					p.buf.WriteString(alt.Type)
				}
				if alt.GroupID != "" {
					p.buf.WriteString(",GROUP-ID=\"")
					p.buf.WriteString(alt.GroupID)
					p.buf.WriteRune('"')
				}
				if alt.Name != "" {
					p.buf.WriteString(",NAME=\"")
					p.buf.WriteString(alt.Name)
					p.buf.WriteRune('"')
				}
				p.buf.WriteString(",DEFAULT=")
				if alt.Default {
					p.buf.WriteString("YES")
				} else {
					p.buf.WriteString("NO")
				}
				if alt.Autoselect != "" {
					p.buf.WriteString(",AUTOSELECT=")
					p.buf.WriteString(alt.Autoselect)
				}
				if alt.Language != "" {
					p.buf.WriteString(",LANGUAGE=\"")
					p.buf.WriteString(alt.Language)
					p.buf.WriteRune('"')
				}
				if alt.Forced != "" {
					p.buf.WriteString(",FORCED=\"")
					p.buf.WriteString(alt.Forced)
					p.buf.WriteRune('"')
				}
				if alt.Characteristics != "" {
					p.buf.WriteString(",CHARACTERISTICS=\"")
					p.buf.WriteString(alt.Characteristics)
					p.buf.WriteRune('"')
				}
				if alt.Subtitles != "" {
					p.buf.WriteString(",SUBTITLES=\"")
					p.buf.WriteString(alt.Subtitles)
					p.buf.WriteRune('"')
				}
				if alt.URI != "" {
					p.buf.WriteString(",URI=\"")
					p.buf.WriteString(alt.URI)
					p.buf.WriteRune('"')
				}
				p.buf.WriteRune('\n')
			}
		}
		if pl.Iframe {
			p.buf.WriteString("#EXT-X-I-FRAME-STREAM-INF:PROGRAM-ID=")
			p.buf.WriteString(strconv.FormatUint(uint64(pl.ProgramID), 10))
			p.buf.WriteString(",BANDWIDTH=")
			p.buf.WriteString(strconv.FormatUint(uint64(pl.Bandwidth), 10))
			if pl.Codecs != "" {
				p.buf.WriteString(",CODECS=\"")
				p.buf.WriteString(pl.Codecs)
				p.buf.WriteRune('"')
			}
			if pl.Resolution != "" {
				p.buf.WriteString(",RESOLUTION=") // Resolution should not be quoted
				p.buf.WriteString(pl.Resolution)
			}
			if pl.Video != "" {
				p.buf.WriteString(",VIDEO=\"")
				p.buf.WriteString(pl.Video)
				p.buf.WriteRune('"')
			}
			if pl.URI != "" {
				p.buf.WriteString(",URI=\"")
				p.buf.WriteString(pl.URI)
				p.buf.WriteRune('"')
			}
			p.buf.WriteRune('\n')
		} else {
			p.buf.WriteString("#EXT-X-STREAM-INF:PROGRAM-ID=")
			p.buf.WriteString(strconv.FormatUint(uint64(pl.ProgramID), 10))
			p.buf.WriteString(",BANDWIDTH=")
			p.buf.WriteString(strconv.FormatUint(uint64(pl.Bandwidth), 10))
			if pl.Codecs != "" {
				p.buf.WriteString(",CODECS=\"")
				p.buf.WriteString(pl.Codecs)
				p.buf.WriteRune('"')
			}
			if pl.Resolution != "" {
				p.buf.WriteString(",RESOLUTION=") // Resolution should not be quoted
				p.buf.WriteString(pl.Resolution)
			}
			if pl.Audio != "" {
				p.buf.WriteString(",AUDIO=\"")
				p.buf.WriteString(pl.Audio)
				p.buf.WriteRune('"')
			}
			if pl.Video != "" {
				p.buf.WriteString(",VIDEO=\"")
				p.buf.WriteString(pl.Video)
				p.buf.WriteRune('"')
			}
			if pl.Captions != "" {
				p.buf.WriteString(",CLOSED-CAPTIONS=")
				if pl.Captions == "NONE" {
					p.buf.WriteString(pl.Captions) // CC should not be quoted when eq NONE
				} else {
					p.buf.WriteRune('"')
					p.buf.WriteString(pl.Captions)
					p.buf.WriteRune('"')
				}
			}
			if pl.Subtitles != "" {
				p.buf.WriteString(",SUBTITLES=\"")
				p.buf.WriteString(pl.Subtitles)
				p.buf.WriteRune('"')
			}
			if pl.Name != "" {
				p.buf.WriteString(",NAME=\"")
				p.buf.WriteString(pl.Name)
				p.buf.WriteRune('"')
			}
			p.buf.WriteRune('\n')
			p.buf.WriteString(pl.URI)
			if p.Args != "" {
				if strings.Contains(pl.URI, "?") {
					p.buf.WriteRune('&')
				} else {
					p.buf.WriteRune('?')
				}
				p.buf.WriteString(p.Args)
			}
			p.buf.WriteRune('\n')
		}
	}

	return &p.buf
}

// Version returns the current playlist version number
func (p *MasterPlaylist) Version() int {
	return p.ver
}

// SetVersion sets the playlist version number, note the version maybe changed
// automatically by other Set methods.
func (p *MasterPlaylist) SetVersion(ver int) {
	p.ver = ver
}

// String returns the encoded buffer in string format,
// which implements the Stringer interface for Printf-like func.
func (p *MasterPlaylist) String() string {
	return p.Encode().String()
}

// NewMediaPlaylist creates a new media playlist structure.
// winsize defines how much items will displayed on playlist generation;
// capacity is total size of a playlist (and its underlying arrray).
func NewMediaPlaylist(winsize, capacity int) (*MediaPlaylist, error) {
	p := new(MediaPlaylist)
	p.ver = minVer
	p.capacity = capacity
	if err := p.SetWinSize(winsize); err != nil {
		return nil, err
	}
	p.Segments = make([]*MediaSegment, capacity)
	return p, nil
}

// last returns the previously written segment's index
func (p *MediaPlaylist) last() int {
	if p.tail == 0 {
		return p.capacity - 1
	}
	return p.tail - 1
}

// Remove current segment from the head of chunk slice form a media playlist. Useful for sliding playlists.
// This operation does reset playlist cache.
func (p *MediaPlaylist) Remove() (err error) {
	if p.count == 0 {
		return errors.New("playlist is empty")
	}
	p.head = (p.head + 1) % p.capacity
	p.count--
	if !p.Closed {
		p.SeqNo++
	}
	p.buf.Reset()
	return nil
}

// Append appends a MediaSegment to the tail of chunk slice for a media playlist.
// This operation does reset playlist cache.
func (p *MediaPlaylist) Append(seg *MediaSegment) error {
	if p.head == p.tail && p.count > 0 {
		return ErrPlaylistFull
	}
	p.Segments[p.tail] = seg
	p.tail = (p.tail + 1) % p.capacity
	p.count++
	if p.TargetDuration < seg.Duration {
		p.TargetDuration = math.Ceil(seg.Duration)
	}
	p.buf.Reset()
	return nil
}

// AppendWithAutoExtend appends a MediaSegment and
// auto extend the capacity if 2/3 full.
func (p *MediaPlaylist) AppendWithAutoExtend(seg *MediaSegment) error {
	if p.count > p.capacity*2/3 {
		if err := p.ExtendCapacity(); err != nil {
			return err
		}
	}
	return p.Append(seg)
}

// Slide first removes one chunk from the head if winsize full, then
// appends one chunk to the tail.
// Useful for sliding/live playlists.
// This operation does reset cache.
func (p *MediaPlaylist) Slide(seg *MediaSegment) (err error) {
	if !p.Closed {
		if p.count >= p.winsize {
			if err = p.Remove(); err != nil {
				return err
			}
		}
		return p.Append(seg)
	}
	return nil
}

// ResetCache resets the playlist cache, so that
// next call on Encode() will regenerate playlist from the chunk slice.
func (p *MediaPlaylist) ResetCache() {
	p.buf.Reset()
}

// Encode generate output in M3U8 format.
// It marshals `winsize` elements from bottom of the `segments` queue.
func (p *MediaPlaylist) Encode() *bytes.Buffer {
	if p.buf.Len() > 0 {
		return &p.buf
	}

	p.buf.WriteString("#EXTM3U\n#EXT-X-VERSION:")
	p.buf.WriteString(strconv.Itoa(p.ver))
	p.buf.WriteRune('\n')
	// default key (workaround for Widevine)
	if p.Key != nil {
		p.buf.WriteString("#EXT-X-KEY:")
		p.buf.WriteString("METHOD=")
		p.buf.WriteString(p.Key.Method)
		if p.Key.Method != "NONE" {
			p.buf.WriteString(",URI=\"")
			p.buf.WriteString(p.Key.URI)
			p.buf.WriteRune('"')
			if p.Key.IV != "" {
				p.buf.WriteString(",IV=")
				p.buf.WriteString(p.Key.IV)
			}
			if p.Key.Keyformat != "" {
				p.buf.WriteString(",KEYFORMAT=\"")
				p.buf.WriteString(p.Key.Keyformat)
				p.buf.WriteRune('"')
			}
			if p.Key.Keyformatversions != "" {
				p.buf.WriteString(",KEYFORMATVERSIONS=\"")
				p.buf.WriteString(p.Key.Keyformatversions)
				p.buf.WriteRune('"')
			}
		}
		p.buf.WriteRune('\n')
	}
	if p.Map != nil {
		p.buf.WriteString("#EXT-X-MAP:")
		p.buf.WriteString("URI=\"")
		p.buf.WriteString(p.Map.URI)
		p.buf.WriteRune('"')
		if p.Map.Limit > 0 {
			p.buf.WriteString(",BYTERANGE=")
			p.buf.WriteString(strconv.Itoa(p.Map.Limit))
			p.buf.WriteRune('@')
			p.buf.WriteString(strconv.Itoa(p.Map.Offset))
		}
		p.buf.WriteRune('\n')
	}
	if p.MediaType > 0 {
		p.buf.WriteString("#EXT-X-PLAYLIST-TYPE:")
		switch p.MediaType {
		case MediaTypeEvent:
			p.buf.WriteString("EVENT\n")
			p.buf.WriteString("#EXT-X-ALLOW-CACHE:NO\n")
		case MediaTypeVOD:
			p.buf.WriteString("VOD\n")
		}
	}
	p.buf.WriteString("#EXT-X-MEDIA-SEQUENCE:")
	p.buf.WriteString(strconv.Itoa(p.SeqNo))
	p.buf.WriteRune('\n')
	p.buf.WriteString("#EXT-X-TARGETDURATION:")
	p.buf.WriteString(strconv.FormatInt(int64(math.Ceil(p.TargetDuration)), 10)) // due section 3.4.2 of M3U8 specs EXT-X-TARGETDURATION must be integer
	p.buf.WriteRune('\n')
	if p.Iframe {
		p.buf.WriteString("#EXT-X-I-FRAMES-ONLY\n")
	}
	// Widevine tags
	if p.W != nil {
		if p.W.AudioChannels != 0 {
			p.buf.WriteString("#WV-AUDIO-CHANNELS ")
			p.buf.WriteString(strconv.FormatUint(uint64(p.W.AudioChannels), 10))
			p.buf.WriteRune('\n')
		}
		if p.W.AudioFormat != 0 {
			p.buf.WriteString("#WV-AUDIO-FORMAT ")
			p.buf.WriteString(strconv.FormatUint(uint64(p.W.AudioFormat), 10))
			p.buf.WriteRune('\n')
		}
		if p.W.AudioProfileIDC != 0 {
			p.buf.WriteString("#WV-AUDIO-PROFILE-IDC ")
			p.buf.WriteString(strconv.FormatUint(uint64(p.W.AudioProfileIDC), 10))
			p.buf.WriteRune('\n')
		}
		if p.W.AudioSampleSize != 0 {
			p.buf.WriteString("#WV-AUDIO-SAMPLE-SIZE ")
			p.buf.WriteString(strconv.FormatUint(uint64(p.W.AudioSampleSize), 10))
			p.buf.WriteRune('\n')
		}
		if p.W.AudioSamplingFrequency != 0 {
			p.buf.WriteString("#WV-AUDIO-SAMPLING-FREQUENCY ")
			p.buf.WriteString(strconv.FormatUint(uint64(p.W.AudioSamplingFrequency), 10))
			p.buf.WriteRune('\n')
		}
		if p.W.CypherVersion != "" {
			p.buf.WriteString("#WV-CYPHER-VERSION ")
			p.buf.WriteString(p.W.CypherVersion)
			p.buf.WriteRune('\n')
		}
		if p.W.ECM != "" {
			p.buf.WriteString("#WV-ECM ")
			p.buf.WriteString(p.W.ECM)
			p.buf.WriteRune('\n')
		}
		if p.W.VideoFormat != 0 {
			p.buf.WriteString("#WV-VIDEO-FORMAT ")
			p.buf.WriteString(strconv.FormatUint(uint64(p.W.VideoFormat), 10))
			p.buf.WriteRune('\n')
		}
		if p.W.VideoFrameRate != 0 {
			p.buf.WriteString("#WV-VIDEO-FRAME-RATE ")
			p.buf.WriteString(strconv.FormatUint(uint64(p.W.VideoFrameRate), 10))
			p.buf.WriteRune('\n')
		}
		if p.W.VideoLevelIDC != 0 {
			p.buf.WriteString("#WV-VIDEO-LEVEL-IDC")
			p.buf.WriteString(strconv.FormatUint(uint64(p.W.VideoLevelIDC), 10))
			p.buf.WriteRune('\n')
		}
		if p.W.VideoProfileIDC != 0 {
			p.buf.WriteString("#WV-VIDEO-PROFILE-IDC ")
			p.buf.WriteString(strconv.FormatUint(uint64(p.W.VideoProfileIDC), 10))
			p.buf.WriteRune('\n')
		}
		if p.W.VideoResolution != "" {
			p.buf.WriteString("#WV-VIDEO-RESOLUTION ")
			p.buf.WriteString(p.W.VideoResolution)
			p.buf.WriteRune('\n')
		}
		if p.W.VideoSAR != "" {
			p.buf.WriteString("#WV-VIDEO-SAR ")
			p.buf.WriteString(p.W.VideoSAR)
			p.buf.WriteRune('\n')
		}
	}

	var (
		seg           *MediaSegment
		durationCache = make(map[float64]string)
	)

	head := p.head
	count := p.count
	for i := 0; (i < p.winsize || p.winsize == 0) && count > 0; count-- {
		seg = p.Segments[head]
		head = (head + 1) % p.capacity
		if seg == nil { // protection from badly filled chunklists
			continue
		}
		if p.winsize > 0 { // skip for VOD playlists, where winsize = 0
			i++
		}
		if seg.SCTE != nil {
			switch seg.SCTE.Syntax {
			case Syntax672014:
				p.buf.WriteString("#EXT-SCTE35:")
				p.buf.WriteString("CUE=\"")
				p.buf.WriteString(seg.SCTE.Cue)
				p.buf.WriteRune('"')
				if seg.SCTE.ID != "" {
					p.buf.WriteString(",ID=\"")
					p.buf.WriteString(seg.SCTE.ID)
					p.buf.WriteRune('"')
				}
				if seg.SCTE.Time != 0 {
					p.buf.WriteString(",TIME=")
					p.buf.WriteString(strconv.FormatFloat(seg.SCTE.Time, 'f', -1, 64))
				}
				p.buf.WriteRune('\n')
			case SyntaxOATCLS:
				switch seg.SCTE.CueType {
				case SCTE35CueStart:
					p.buf.WriteString("#EXT-OATCLS-SCTE35:")
					p.buf.WriteString(seg.SCTE.Cue)
					p.buf.WriteRune('\n')
					p.buf.WriteString("#EXT-X-CUE-OUT:")
					p.buf.WriteString(strconv.FormatFloat(seg.SCTE.Time, 'f', -1, 64))
					p.buf.WriteRune('\n')
				case SCTE35CueMid:
					p.buf.WriteString("#EXT-X-CUE-OUT-CONT:")
					p.buf.WriteString("ElapsedTime=")
					p.buf.WriteString(strconv.FormatFloat(seg.SCTE.Elapsed, 'f', -1, 64))
					p.buf.WriteString(",Duration=")
					p.buf.WriteString(strconv.FormatFloat(seg.SCTE.Time, 'f', -1, 64))
					p.buf.WriteString(",SCTE35=")
					p.buf.WriteString(seg.SCTE.Cue)
					p.buf.WriteRune('\n')
				case SCTE35CueEnd:
					p.buf.WriteString("#EXT-X-CUE-IN")
					p.buf.WriteRune('\n')
				}
			}
		}
		// check for key change
		if seg.Key != nil && p.Key != seg.Key {
			p.buf.WriteString("#EXT-X-KEY:")
			p.buf.WriteString("METHOD=")
			p.buf.WriteString(seg.Key.Method)
			if seg.Key.Method != "NONE" {
				p.buf.WriteString(",URI=\"")
				p.buf.WriteString(seg.Key.URI)
				p.buf.WriteRune('"')
				if seg.Key.IV != "" {
					p.buf.WriteString(",IV=")
					p.buf.WriteString(seg.Key.IV)
				}
				if seg.Key.Keyformat != "" {
					p.buf.WriteString(",KEYFORMAT=\"")
					p.buf.WriteString(seg.Key.Keyformat)
					p.buf.WriteRune('"')
				}
				if seg.Key.Keyformatversions != "" {
					p.buf.WriteString(",KEYFORMATVERSIONS=\"")
					p.buf.WriteString(seg.Key.Keyformatversions)
					p.buf.WriteRune('"')
				}
			}
			p.buf.WriteRune('\n')
		}
		if seg.Discontinuity {
			p.buf.WriteString("#EXT-X-DISCONTINUITY\n")
		}
		// ignore segment Map if default playlist Map is present
		if p.Map == nil && seg.Map != nil {
			p.buf.WriteString("#EXT-X-MAP:")
			p.buf.WriteString("URI=\"")
			p.buf.WriteString(seg.Map.URI)
			p.buf.WriteRune('"')
			if seg.Map.Limit > 0 {
				p.buf.WriteString(",BYTERANGE=")
				p.buf.WriteString(strconv.Itoa(seg.Map.Limit))
				p.buf.WriteRune('@')
				p.buf.WriteString(strconv.Itoa(seg.Map.Offset))
			}
			p.buf.WriteRune('\n')
		}
		if !seg.ProgramDateTime.IsZero() {
			p.buf.WriteString("#EXT-X-PROGRAM-DATE-TIME:")
			p.buf.WriteString(seg.ProgramDateTime.Format(DateTime))
			p.buf.WriteRune('\n')
		}
		if seg.Limit > 0 {
			p.buf.WriteString("#EXT-X-BYTERANGE:")
			p.buf.WriteString(strconv.Itoa(seg.Limit))
			p.buf.WriteRune('@')
			p.buf.WriteString(strconv.Itoa(seg.Offset))
			p.buf.WriteRune('\n')
		}
		p.buf.WriteString("#EXTINF:")
		if str, ok := durationCache[seg.Duration]; ok {
			p.buf.WriteString(str)
		} else {
			if p.durationAsInt {
				// Old Android players has problems with non integer Duration.
				durationCache[seg.Duration] = strconv.FormatInt(int64(math.Ceil(seg.Duration)), 10)
			} else {
				// Wowza Mediaserver and some others prefer floats.
				durationCache[seg.Duration] = strconv.FormatFloat(seg.Duration, 'f', 3, 32)
			}
			p.buf.WriteString(durationCache[seg.Duration])
		}
		p.buf.WriteRune(',')
		p.buf.WriteString(seg.Title)
		p.buf.WriteRune('\n')
		p.buf.WriteString(seg.URI)
		if p.Args != "" {
			p.buf.WriteRune('?')
			p.buf.WriteString(p.Args)
		}
		p.buf.WriteRune('\n')
	}
	if p.Closed {
		p.buf.WriteString("#EXT-X-ENDLIST\n")
	}
	return &p.buf
}

// String returns the encoded buffer in string format,
// which implements the Stringer interface for Printf-like func.
func (p *MediaPlaylist) String() string {
	return p.Encode().String()
}

// DurationAsInt sets if the TargetDuration should be format to int.
func (p *MediaPlaylist) DurationAsInt(yes bool) {
	if yes {
		// duration must be integers if protocol version is less than 3
		checkVersion(&p.ver, 3)
	}
	p.durationAsInt = yes
}

// Count returns the number of items that are currently in the media playlist.
func (p *MediaPlaylist) Count() int {
	return p.count
}

// Close adds end-list to the playlist.
func (p *MediaPlaylist) Close() {
	if p.buf.Len() > 0 {
		p.buf.WriteString("#EXT-X-ENDLIST\n")
	}
	p.Closed = true
}

// SetDefaultKey sets the encryption key appeared once in header of the playlist
// (pointer to MediaPlaylist.Key).
// The tag set applies for the whole list.
// It is useful when keys are not changed during playback.
func (p *MediaPlaylist) SetDefaultKey(method, uri, iv, keyformat, keyformatversions string) error {
	// A Media Playlist MUST indicate a EXT-X-VERSION of 5 or higher if it
	// contains:
	//   - The KEYFORMAT and KEYFORMATVERSIONS attributes of the EXT-X-KEY tag.
	if keyformat != "" || keyformatversions != "" {
		checkVersion(&p.ver, 5)
	}
	p.Key = &Key{method, uri, iv, keyformat, keyformatversions}

	return nil
}

// SetDefaultMap sets the default Media Initialization Section values
// for playlist (pointer to MediaPlaylist.Map).
// It sets EXT-X-MAP tag for the whole playlist.
func (p *MediaPlaylist) SetDefaultMap(uri string, limit, offset int) {
	checkVersion(&p.ver, 5) // due section 4
	p.Map = &Map{uri, limit, offset}
}

// SetIframeOnly marks medialist as consists of only I-frames (Intra frames).
// The tag set applies for the whole list.
func (p *MediaPlaylist) SetIframeOnly() {
	checkVersion(&p.ver, 4) // due section 4.3.3
	p.Iframe = true
}

// SetKey sets a encryption key for the current segment of media playlist
// (pointer to Segment.Key).
func (p *MediaPlaylist) SetKey(method, uri, iv, keyformat, keyformatversions string) error {
	if p.count == 0 {
		return errors.New("playlist is empty")
	}

	// A Media Playlist MUST indicate a EXT-X-VERSION of 5 or higher if it
	// contains:
	//   - The KEYFORMAT and KEYFORMATVERSIONS attributes of the EXT-X-KEY tag.
	if keyformat != "" || keyformatversions != "" {
		checkVersion(&p.ver, 5)
	}

	p.Segments[p.last()].Key = &Key{method, uri, iv, keyformat, keyformatversions}
	return nil
}

// SetMap sets a map for the current segment of media playlist
// (pointer to Segment.Map).
func (p *MediaPlaylist) SetMap(uri string, limit, offset int) error {
	if p.count == 0 {
		return errors.New("playlist is empty")
	}
	checkVersion(&p.ver, 5) // due section 4
	p.Segments[p.last()].Map = &Map{uri, limit, offset}
	return nil
}

// SetRange sets the limit and offset for the current media segment
// (EXT-X-BYTERANGE support for protocol version 4).
func (p *MediaPlaylist) SetRange(limit, offset int) error {
	if p.count == 0 {
		return errors.New("playlist is empty")
	}
	checkVersion(&p.ver, 4) // due section 3.4.1
	p.Segments[p.last()].Limit = limit
	p.Segments[p.last()].Offset = offset
	return nil
}

// SetSCTE35 sets the SCTE cue format for the current media segment
func (p *MediaPlaylist) SetSCTE35(scte35 *SCTE) error {
	if p.count == 0 {
		return errors.New("playlist is empty")
	}
	p.Segments[p.last()].SCTE = scte35
	return nil
}

// SetDiscontinuity sets the discontinuity-flag for the current media segment.
// EXT-X-DISCONTINUITY indicates an encoding discontinuity between the media segment
// that follows it and the one that preceded it (i.e. file format, number and type of tracks,
// encoding parameters, encoding sequence, timestamp sequence).
func (p *MediaPlaylist) SetDiscontinuity() error {
	if p.count == 0 {
		return errors.New("playlist is empty")
	}
	p.Segments[p.last()].Discontinuity = true
	return nil
}

// SetProgramDateTime sets the program date and time for the current media segment.
// EXT-X-PROGRAM-DATE-TIME tag associates the first sample of a
// media segment with an absolute date and/or time.  It applies only
// to the current media segment.
// Date/time format is YYYY-MM-DDThh:mm:ssZ (ISO8601) and includes time zone.
func (p *MediaPlaylist) SetProgramDateTime(value time.Time) error {
	if p.count == 0 {
		return errors.New("playlist is empty")
	}
	p.Segments[p.last()].ProgramDateTime = value
	return nil
}

// Version returns the current playlist version number
func (p *MediaPlaylist) Version() int {
	return p.ver
}

// SetVersion sets the playlist version number, note the version maybe changed
// automatically by other Set methods.
func (p *MediaPlaylist) SetVersion(ver int) {
	p.ver = ver
}

// WinSize returns the playlist's window size.
func (p *MediaPlaylist) WinSize() int {
	return p.winsize
}

// SetWinSize overwrites the playlist's window size.
func (p *MediaPlaylist) SetWinSize(winsize int) error {
	if winsize > p.capacity {
		return errors.New("capacity must be greater than winsize or equal")
	}
	p.winsize = winsize
	return nil
}

// ExtendCapacity extend current capacity to double, and move tail to pos of last end
// When Generate Vod Playlist, if append failed, can call this extend function to contine
func (p *MediaPlaylist) ExtendCapacity() (err error) {
	if p.count == 0 {
		return errors.New("when extend capcity, cur count should > 0")
	}
	p.Segments = append(p.Segments, make([]*MediaSegment, p.count)...)
	p.capacity = len(p.Segments)
	p.tail = p.count
	return
}
