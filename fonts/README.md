# Local typefaces

- IBM Plex Sans Regular / SemiBold and Plex Mono Regular / Italic: Latin-1 WOFF2
  subsets from the official [IBM Plex repository](https://github.com/IBM/plex),
  `packages/plex-{sans,mono}/fonts/split/woff2/`. License: `IBM-Plex-LICENSE.txt`.

The declarations live in `static/css/site.css`. No remote font service is required.

- Source Serif 4 Regular / SemiBold / Bold / Italic: self-hosted TTF files from the
  Google Fonts distribution of [Adobe Source Serif](https://github.com/adobe-fonts/source-serif).
  License: `Source-Serif-LICENSE.md`. Used for prose and headings in
  `static/css/site.css`; IBM Plex remains on navigation and controls.

- Yinit ornamental initials: OpenType conversion from [CTAN](https://ctan.org/pkg/yinit-otf),
  original by Yannis Haralambous, public domain. Attribution and conversion notes:
  `Yinit-README.md`. Retained as source material; no longer used or published.

Source Serif 4 is served as Latin variable WOFF2 (weights 400–700, optical sizes
8–60), using the official Google Fonts distribution of Adobe's typeface. The
body uses a real 450 weight. The unused Yinit sources remain alongside their license.

EB Garamond Roman and Italic: Latin variable WOFF2, weights 400–700, from
Google Fonts' EB Garamond distribution. SIL Open Font License in
`EB-Garamond-OFL.txt`. Retained from the typography trial; no longer used or published.
