Completely unorganized and without a deep thought behind.

- [x] Leader key to trigger commands similar to nvim?
- [x] Initialization screen supports j/k to move up/down the log lines.
- [X] Metatag parsing of audio files.
- [X] Organization and definitions of albums and music, based on the retrieved metatag.
- [X] Cache of previously scanned files.
- [X] "Scan new files" action: appends any newly found files to the existing cache instead of a full rescan.
- [X] Dialog on the screen which *can* trigger actions depending on user's response.
- [x] Decide what cards to show on librarystats depending on window size.
- [x] Refactor `assertSnapshot` and the `-update` golden flag into shared `testutil`.
- [x] Notification system to send messages to the user. (especially useful for debugging)
- [ ] View: Application layout design.
- [ ] View: Currently library.
- [ ] View: Currently selected album.
- [ ] View: Currently selected artist.
- [ ] View: Currently playing playlist.
- [ ] View: Currently playing track visualization, including spectrum, cover if any, lyrics if any.
- [ ] Last.fm scrobbling.
- [ ] Radio integration.
- [ ] Metatag editing in librarystats view.
- [ ] Librarystats dump to a file.
- [ ] Dedicated logger purely for the file discovery/parsing pipeline (parse errors are currently swallowed and nothing is logged).
	- [ ] Maybe it's better to have a dedicated logger service as a core to be used in many different places?

? - Maybe this is a bad idea. Needs more thought.

