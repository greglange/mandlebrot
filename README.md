# Mandlebrot set

Using this project you can make images and videos of the Mandlebrot set.

## Zoom video

[This](https://www.youtube.com/watch?v=Z5R7WNN8Hbs) video was made using this project.

## Building the project

Clone or download this project and change dir into the repo.

Then run:

`go install cmd/mandlebrot.go`

## Making an image

Run:

`./image`

An image should be written to `image.png`.

## Making a video

To make a video you need to have [ffmpeg](https://ffmpeg.org/) or similar video encoder installed.

To make the video from the link above, run:

`./video`

Your video should be written to `video.mp4`.
