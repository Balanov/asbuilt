# The paritet.dev block

The block that asbuilt gets on the paritet.dev home page lives here, next to
the page source it is spliced into.

- `asbuilt-evidence-template.html` — the template, in the same skeleton as the
  Skein block beside it: kicker, headline, facts on the left, the measured
  numbers on the right, environment and links below.
- `paritet-asbuilt.css` — the styles, scoped to `#asbuilt-evidence` and the
  shared slider.
- `paritet-asbuilt.js` — inserts the block, keeps it next to the Skein block,
  and drives the slider: Skein first, asbuilt a few seconds later, with tabs
  for both.
- `patch_block.py` — splices the template into the home page between its
  markers, so the page can be newer than this copy without being overwritten.

The home page is owned by the main site pipeline. The block is therefore
patched in place rather than deployed as a whole page:

```sh
cp asbuilt-evidence-template.html patch_block.py /tmp/homepatch/
cd /tmp/homepatch && python3 patch_block.py ~/paritet/public_html/index.html
```

The template sits between two markers:

```html
<!-- asbuilt:block:start -->
...
<!-- asbuilt:block:end -->
```

The script replaces what is between the markers, and on the first run wraps
the plain `<template>` instead. Run it again after changing the template.
