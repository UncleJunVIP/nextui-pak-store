<div align="center">
<img src=".github/resources/banner.png" width="auto" alt="Mortar wordmark">

![GitHub License](https://img.shields.io/github/license/UncleJunVip/nextui-pak-store?style=for-the-badge&color=007C77)
![GitHub Release](https://img.shields.io/github/v/release/UncleJunVIP/nextui-pak-store?sort=semver&style=for-the-badge&color=007C77)
![GitHub Repo stars](https://img.shields.io/github/stars/UncleJunVip/nextui-pak-store?style=for-the-badge&color=007C77)
![GitHub Downloads (specific asset, all releases)](https://img.shields.io/github/downloads/UncleJunVIP/nextui-pak-store/total?style=for-the-badge&label=Downloads&color=007C77)

</div>

---

## How do I setup Pak Store?

1. Own a TrimUI Brick or Smart Pro and have a SD Card with NextUI configured.
2. Connect your device to a Wi-Fi network.
3. Download the latest Pak Store release from this repo.
4. Unzip the release download. 
   - If the unzipped folder name is `Pak.Store.pak` please rename it to `Pak Store.pak`.
5. Copy the entire `Pak Store.pak` folder to `SD_ROOT/Tools/tg5040`.
6. Reinsert your SD Card into your device.
7. Launch `Pak Store` from the `Tools` menu and enjoy all the amazing Paks made by the community!

---

## I want my Pak in Pak Store!

Awesome! To get added to Pak Store you have to complete the following steps:

1. Create a `pak.json` file at the root of your repo. An example can be seen below.
   - The following fields are **required**
     - `name`
     - `version`
     - `type`
     - `description`
     - `author`
     - `repo_url`
     - `release_filename`
     - `platforms`
   - If you are packaging up an emulator, please set the name to the desired emulator tag. (e.g., an Intellivision Pak with the tag `INTV` would have `INTV` as the name in pak.json)
2. Prepare your Pak for distribution by making a zip file. The contents of the zip file must the contents present in the root of your Pak directory.
3. Ensure your release is tagged properly and matches the `version` field in `pak.json`.
   - The tag should be in the format `vX.X.X` where `X` is the major, minor, and patch version. For more details for using SemVer, please see the [SemVer Documentation](https://semver.org/).
   - GitHub releases have both tags and titles. The title does not matter in the context of the Pak Store but you should have it match the tag and pak.json version.
4. Make sure the file name of the release artifact matches what is in `pak.json`.
5. Once all of these steps are complete, please file an issue with a link to your repo.

---

## Sample pak.json
```json
{
  "name": "Pak Store",
  "version": "v1.0.1",
  "type": "TOOL | EMU",
  "description": "A Pak Store in this economy?!",
  "author": "K-Wall",
  "repo_url": "https://github.com/UncleJunVIP/nextui-pak-store",
  "release_filename": "Pak.Store.pak.zip",
  "changelog": {
    "v1.0.0": "Upgraded the UI to use gabagool, my NextUI Pak UI Library!"
  },
  "update_ignore": [
    "path/of/file/to/ignore"
  ],
  "screenshots": [
    ".github/resources/screenshots/main_menu.jpg",
    ".github/resources/screenshots/browse.jpg",
    ".github/resources/screenshots/ports.jpg",
    ".github/resources/screenshots/portmaster_1.jpg",
    ".github/resources/screenshots/portmaster_2.jpg",
    ".github/resources/screenshots/updates.jpg"
  ],
  "platforms": [
    "tg5040"
  ]
}
```

---

Enjoy! ✌🏻