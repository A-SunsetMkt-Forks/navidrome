package metadata

import (
	"encoding/json"
	"math"
	"strconv"

	"github.com/navidrome/navidrome/log"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/utils/str"
)

func (md Metadata) ToMediaFile(libID int, folderID string) model.MediaFile {
	mf := model.MediaFile{
		LibraryID: libID,
		FolderID:  folderID,
		Tags:      md.tags,
	}

	// Title and Album
	mf.Title = md.mapTrackTitle()
	mf.Album = md.mapAlbumName()
	mf.SortTitle = md.String(model.TagTitleSort)
	mf.SortAlbumName = md.String(model.TagAlbumSort)
	mf.OrderTitle = str.SanitizeFieldForSorting(mf.Title)
	mf.OrderAlbumName = str.SanitizeFieldForSortingNoArticle(mf.Album)
	mf.Compilation = md.Bool(model.TagCompilation)

	// Disc and Track info
	mf.TrackNumber, _ = md.NumAndTotal(model.TagTrackNumber)
	mf.DiscNumber, _ = md.NumAndTotal(model.TagDiscNumber)
	mf.DiscSubtitle = md.String(model.TagDiscSubtitle)
	mf.CatalogNum = md.String(model.TagCatalogNumber)
	mf.Comment = md.String(model.TagComment)
	mf.BPM = int(math.Round(md.Float(model.TagBPM)))
	mf.Lyrics = md.mapLyrics()

	// Dates
	origDate := md.Date(model.TagOriginalDate)
	mf.OriginalYear, mf.OriginalDate = origDate.Year(), string(origDate)
	relDate := md.Date(model.TagReleaseDate)
	mf.ReleaseYear, mf.ReleaseDate = relDate.Year(), string(relDate)
	date := md.Date(model.TagRecordingDate)
	mf.Year, mf.Date = date.Year(), string(date)

	// MBIDs
	mf.MbzRecordingID = md.String(model.TagMusicBrainzRecordingID)
	mf.MbzReleaseTrackID = md.String(model.TagMusicBrainzTrackID)
	mf.MbzAlbumID = md.String(model.TagMusicBrainzAlbumID)
	mf.MbzReleaseGroupID = md.String(model.TagMusicBrainzReleaseGroupID)

	// ReplayGain
	mf.RGAlbumPeak = md.Float(model.TagReplayGainAlbumPeak, 1)
	mf.RGAlbumGain = md.mapGain(model.TagReplayGainAlbumGain, model.TagR128AlbumGain)
	mf.RGTrackPeak = md.Float(model.TagReplayGainTrackPeak, 1)
	mf.RGTrackGain = md.mapGain(model.TagReplayGainTrackGain, model.TagR128TrackGain)

	// General properties
	mf.HasCoverArt = md.HasPicture()
	mf.Duration = md.Length()
	mf.BitRate = md.AudioProperties().BitRate
	mf.SampleRate = md.AudioProperties().SampleRate
	mf.BitDepth = md.AudioProperties().BitDepth
	mf.Channels = md.AudioProperties().Channels
	mf.Path = md.FilePath()
	mf.Suffix = md.Suffix()
	mf.Size = md.Size()
	mf.BirthTime = md.BirthTime()
	mf.UpdatedAt = md.ModTime()

	// BFR Parse and import lyrics
	//mf.Lyrics = md.String(model.TagLyrics)

	mf.Participations = md.mapParticipations()
	mf.Artist = md.mapDisplayArtist(mf)
	mf.AlbumArtist = md.mapDisplayAlbumArtist(mf)

	// Persistent IDs
	mf.PID = md.trackPID(mf)
	mf.AlbumID = md.albumID(mf)

	// BFR These IDs will go away once the UI handle multiple participants.
	// BFR For Legacy Subsonic compatibility, we will set them in the API handlers
	mf.ArtistID = mf.Participations.First(model.RoleArtist).ID
	mf.AlbumArtistID = mf.Participations.First(model.RoleAlbumArtist).ID

	// BFR What to do with sort/order artist names?
	mf.OrderArtistName = mf.Participations.First(model.RoleArtist).OrderArtistName
	mf.OrderAlbumArtistName = mf.Participations.First(model.RoleAlbumArtist).OrderArtistName
	mf.SortArtistName = mf.Participations.First(model.RoleArtist).SortArtistName
	mf.SortAlbumArtistName = mf.Participations.First(model.RoleAlbumArtist).SortArtistName

	// Don't store tags that are first-class fields in the MediaFile struct
	// BFR Automatically remove tags that were accessed in the steps above
	tagsToIgnore := []model.TagName{
		model.TagAlbum, model.TagTitle, model.TagTrackNumber, model.TagDiscNumber, model.TagDiscSubtitle,
		model.TagComment, model.TagAlbumSort, model.TagTitleSort, model.TagCompilation,
		model.TagLyrics, model.TagCatalogNumber, model.TagBPM, model.TagOriginalDate,
		model.TagReleaseDate, model.TagRecordingDate,

		// MusicBrainz IDs
		model.TagMusicBrainzRecordingID, model.TagMusicBrainzTrackID, model.TagMusicBrainzAlbumID,
		model.TagMusicBrainzReleaseGroupID, model.TagMusicBrainzAlbumArtistID, model.TagMusicBrainzArtistID,

		// ReplayGain
		model.TagReplayGainAlbumPeak, model.TagReplayGainAlbumGain, model.TagReplayGainTrackPeak, model.TagReplayGainTrackGain,
		model.TagR128AlbumGain, model.TagR128TrackGain,

		// Roles
		model.TagComposer, model.TagConductor, model.TagArranger, model.TagLyricist, model.TagRemixer,
		model.TagEngineer, model.TagMixer, model.TagProducer, model.TagDirector, model.TagDJMixer,
		model.TagPerformer, model.TagAlbumArtist, model.TagAlbumArtists, model.TagAlbumArtistSort,
		model.TagAlbumArtistsSort, model.TagTrackArtist, model.TagTrackArtists, model.TagTrackArtistSort,
		model.TagTrackArtistsSort, model.TagComposerSort,
	}
	for _, tag := range tagsToIgnore {
		delete(mf.Tags, tag)
	}

	return mf
}

func (md Metadata) mapGain(rg, r128 model.TagName) float64 {
	v := md.Gain(rg)
	if v != 0 {
		return v
	}
	r128value := md.String(r128)
	if r128value != "" {
		var v, err = strconv.Atoi(r128value)
		if err != nil {
			return 0
		}
		// Convert Q7.8 to float
		var value = float64(v) / 256.0
		// Adding 5 dB to normalize with ReplayGain level
		return value + 5
	}
	return 0
}

func (md Metadata) mapLyrics() string {
	rawLyrics := md.Pairs(model.TagLyrics)

	lyricList := make(model.LyricList, 0, len(rawLyrics))

	for _, raw := range rawLyrics {
		lang := raw.Key()
		text := raw.Value()

		lyrics, err := model.ToLyrics(lang, text)
		if err != nil {
			log.Warn("Unexpected failure occurred when parsing lyrics", "file", md.filePath, err)
			continue
		}
		lyricList = append(lyricList, *lyrics)
	}

	res, err := json.Marshal(lyricList)
	if err != nil {
		log.Warn("Unexpected error occurred when serializing lyrics", "file", md.filePath, err)
		return ""
	}
	return string(res)
}
