package metadata

import (
	"testing"

	"github.com/navidrome/navidrome/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMetadata(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metadata Suite")
}

var _ = Describe("getPID", func() {
	var (
		md     Metadata
		sum    hashFunc
		getPID func(md Metadata, spec string) string
	)

	BeforeEach(func() {
		sum = func(s string) string { return s }
		getPID = createGetPID(sum)
	})

	Context("attributes are tags", func() {
		spec := "musicbrainz_trackid|album,discnumber,tracknumber,title"
		When("no attributes were present", func() {
			It("should return empty pid", func() {
				md.tags = map[model.TagName][]string{}
				pid := getPID(md, spec)
				Expect(pid).To(BeEmpty())
			})
		})
		When("all fields are present", func() {
			It("should return the pid", func() {
				md.tags = map[model.TagName][]string{
					"musicbrainz_trackid": {"mbtrackid"},
					"album":               {"albumname"},
					"discnumber":          {"1"},
					"tracknumber":         {"1"},
					"title":               {"song title"},
				}
				Expect(getPID(md, spec)).To(Equal("mbtrackid"))
			})
		})
		When("only first field is present", func() {
			It("should return the pid", func() {
				md.tags = map[model.TagName][]string{
					"musicbrainz_trackid": {"mbtrackid"},
				}
				Expect(getPID(md, spec)).To(Equal("mbtrackid"))
			})
		})
		When("first is empty, but second field is present", func() {
			It("should return the pid", func() {
				md.tags = map[model.TagName][]string{
					"album":      {"albumname"},
					"discnumber": {"1"},
					"title":      {"song title"},
				}
				Expect(getPID(md, spec)).To(Equal("albumname\\1\\\\song title"))
			})
		})
	})
	Context("calculated attributes", func() {
		When("field is folder", func() {
			It("should return the pid", func() {
				spec := "folder|title"
				md.tags = map[model.TagName][]string{"title": {"title"}}
				md.filePath = "/path/to/file.mp3"
				Expect(getPID(md, spec)).To(Equal("/path/to"))
			})
		})
		When("field is albumid", func() {
			It("should return the pid", func() {
				spec := "albumid|title"
				md.tags = map[model.TagName][]string{
					"title":       {"title"},
					"album":       {"albumname"},
					"version":     {"version"},
					"releasedate": {"2021-01-01"},
				}
				Expect(getPID(md, spec)).To(Equal("\\albumname\\version\\2021-01-01"))
			})
		})
	})
})
