# -*- coding: utf-8 -*-
"""Splice the asbuilt block into the paritet home page.

Runs on the server: the home page is owned by the main site pipeline and may
already be newer than the local copy, so the block is replaced between its
markers instead of rewriting the file.
"""
import pathlib
import re
import sys

HERE = pathlib.Path(__file__).resolve().parent
TEMPLATE = HERE / "asbuilt-template.html"
START = "<!-- asbuilt:block:start -->"
END = "<!-- asbuilt:block:end -->"


def main():
    target = pathlib.Path(sys.argv[1] if len(sys.argv) > 1 else "index.html")
    text = target.read_text(encoding="utf-8")
    body = TEMPLATE.read_text(encoding="utf-8")
    block = START + "\n" + body + "    " + END

    pattern = re.compile(re.escape(START) + r".*?" + re.escape(END), re.S)
    if pattern.search(text):
        updated = pattern.sub(lambda _: block, text, count=1)
        action = "replaced"
    else:
        pattern = re.compile(r"(\s*)<template id=\"asbuilt-evidence-template\">.*?</template>", re.S)
        match = pattern.search(text)
        if not match:
            print("asbuilt block not found")
            return 1
        updated = text[:match.start()] + match.group(1) + block + text[match.end():]
        action = "wrapped"

    target.write_text(updated, encoding="utf-8")
    print(action, len(updated))
    return 0


if __name__ == "__main__":
    sys.exit(main())
