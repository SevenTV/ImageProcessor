# ImageConverter

The emote converter is a microservice used to convert uploaded raw files to the 7TV emote format.
There are 3 stages to an emote upload.

## Supported Upload Types

| Format      | Supports Animation | Supports Transparency |
| :---:       | :---:              | :---:                 |
|  AVI        | ✅¹               | ❌                    |
|  AVIF       | ✅ ​               | ✅                    |
|  FLV        | ✅¹               | ❌                    |
|  GIF        | ✅   ​   ​          | ✅                    |
|  JPEG       | ❌   ​  ​           | ❌                    |
|  MP4        | ✅¹               | ❌                    |
|  MOV        |​​​​ ✅¹               ​| ❌                    |
|  PNG/APNG   | ✅¹               | ✅                    |
|  TIFF       | ❌ ​ ​              | ✅                    |
|  WEBM       | ✅¹               ​| ❌                    |
|  WEBP       |​​​​ ✅​​   ​        ​     ​| ✅                    |

1. Emotes uploaded in these formats will not have any audio and per frame delays will not be considered. The frame rate of the video will be used to decide the delay per frame.

    Which means `frame_delay = 1000 / frames_per_second`

## Stages of an emote upload

### Stage 1

Animated Emotes

    If the emote is animated meaning:

    1. The emote is uploaded using one of the supported animated containers.
    2. The emote has more than one frame.

    Then the emote is converted to a series of PNG images with the respective delays attached.

Static Emotes

    If the emote is static meaning:

    1. It is uploaded in a static only format.
    2. It is uploaded in a animation format but only has one frame.

    Then the emote will be converted to a single PNG file.

### Stage 2

All PNG images are resized to be with in the following size ranges defined by the job payload.

The default size ranges are:

| Name  | Height | Minimum Width | Maximum Width |
| :---: | :----: | :-----------: | :-----------: |
| 1x    | 32px   | 32px          | 96px          |
| 2x    | 64px   | 64px          | 192px         |
| 3x    | 96px   | 96px          | 288px         |
| 4x    | 128px  | 128px         | 384px         |

We DO NOT crop images, but rather resize them.

- If the ratio of width to height is greater than 3:1 we then must first shrink to the acceptable width and then we place the emote in the bottom left corner and make sure the canvas has a ratio of 3:1. For example, if we recieve an emote that is 6:1 in size, lets say 768px in width by 128px in height. We will first shrink this emote respecting the aspect ratio to 386px in width by 64px in height, then we will add 64px of transparent pixels to the top of the emote to make it 384px by 128px.

- If the emote has a height to width ratio less than 1:1 then we resize the canvas to make it 1:1. For example, if an emote is uploaded with a height of 128px and a width of 64px then we will add 64px of transparent pixels to the right part of the image to make the canvas size 128px in height  by 128px width.

These 2 rules govern the outcome of emotes which are uploaded. This step will create 4 size variants.

### Stage 3

Animated Emotes

    If the emote is considered animated from stage 1
    Then we must convert the emote to an animated WEBP, GIF, AVIF with all size variants which are 1x, 2x, 3x, and 4x.
    Animated emotes have 3 type variants being WEBP, GIF and AVIF.
    We will then also take the first frame of the animated emote and convert it to WEBP and AVIF for a thumbnail of the emote. We then reuse the PNG from stage 2 to make a thumbnail which contains all 3 variants.
    In total this results in 24 images, 4 size variants images per image type and then multipled by 2 for thumbnails.
    24 = 4 * 3 * 2 

Static Emotes

    If the emote is considered static from stage 1
    Then we will only convert it to WEBP and AVIF.
    We will reuse the PNG from stage 2 to also have 3 variants for static emotes.
    Static emotes have 3 type variants being WEBP, PNG and AVIF with all size variants which are 1x, 2x, 3x, and 4x.
    This will result in 12 images, 4 size variants images per image type.
