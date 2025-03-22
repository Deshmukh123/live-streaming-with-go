document.addEventListener("DOMContentLoaded", function () {
    var video = document.getElementById("videoPlayer");
    var videoSrc = "videos/stream.m3u8";

    if (Hls.isSupported()) {
        var hls = new Hls();
        hls.loadSource(videoSrc);
        hls.attachMedia(video);
        hls.on(Hls.Events.MANIFEST_PARSED, function () {
            video.play();
        });
    } else if (video.canPlayType("application/vnd.apple.mpegurl")) {
        video.src = videoSrc;
        video.play();
    }

    function fetchVideoDetails() {
        fetch("/video-details")
            .then(response => response.json())
            .then(data => {
                document.getElementById("videoDetails").innerHTML =
                    `Resolution: ${data.resolution} | Duration: ${data.duration} | Size: ${data.size}`;
            })
            .catch(error => console.error("Error fetching details:", error));
    }

    setInterval(fetchVideoDetails, 5000); // Refresh every 5 seconds
});
