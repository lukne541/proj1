
class Serie {
  constructor(seasons, episodes_max, episodes_min) {
    this.episodes_max = episodes_max;
    this.episodes_min = episodes_min;
    this.seasons = seasons;
  }
}
var series = {};

fetch("/get_titles")
.then(response => {
  return response.json();
})
.then(jsonResponse => {
  var x = document.getElementsByName("serie")[0];
  var y = document.getElementsByName("season")[0];
  var z = document.getElementsByName("episode")[0];
  var titles = jsonResponse.Titles;
  for (var i = 0; i < titles.length; i++) {
    var option = document.createElement("option");
    var j = titles[i];
    option.value = j.Title;
    option.text = j.Title;
    x.add(option);
    series[j.Title] = new Serie(j.Seasons, j.Episodes_max, j.Episodes_min);
    updateForm();
    
  }
});

var x = document.getElementsByName("serie")[0].addEventListener("change", updateForm);

function updateForm() {

  var x = document.getElementsByName("serie")[0];
  var y = document.getElementsByName("season")[0];

  y.options.length = 0;

  var serie = series[x.options[x.selectedIndex].value];

  for (var i = 1; i < serie.seasons + 1; i++) {
    var opt = document.createElement("option");
    opt.value = i;
    opt.text = i;
    y.add(opt);
  }
  updateEpisodeCount();
}

var form1 = document.querySelector(".my-form");

form1.addEventListener("submit", e => {
  playVideo();
});


function updateEpisodeCount() {
  var x = document.getElementsByName("serie")[0];
  var select_e = document.getElementsByName("episode")[0];
  var select_s = document.getElementsByName("season")[0];
  
  var season = select_s.options[select_s.selectedIndex].value;
  var serie = series[x.options[x.selectedIndex].value];

  select_e.options.length = 0;

  for (var i = serie.episodes_min[season - 1] + 1; i <= serie.episodes_max[season - 1] + 1; i++) {
    var opt = document.createElement("option");
    opt.value = i - serie.episodes_min[season - 1];
    opt.text = i - serie.episodes_min[season - 1];
    select_e.add(opt);
  }
}

document.getElementsByName("season")[0].addEventListener("change", updateEpisodeCount);

function playVideo() {
  var formData = new FormData(document.getElementById('form1'));
  //window.alert(formData.get("serie"));
  var serie = formData.get("serie");
  var season = formData.get("season");
  var episode = +formData.get("episode") + +series[serie].episodes_min[season - 1] - 1;
  
  // Simulate an HTTP redirect:
  fetch("/watch/" + serie + "/" + season + "/" + episode, {method: 'POST'}).then(response =>  {
    // TODO: Webserver will send json object with url to track (if exists) and media
    return response.json();
  })
  .then(jsonResponse => {
    var video;
    var firstBool = true;

    videoDiv = document.getElementById("videoDiv");
    video = document.createElement("video");

    video.src = jsonResponse.VidUrl;
    video.autolay = true;
    video.setAttribute("controls","controls")
    video.id = "videoPlayer";
    

    video.addEventListener("loadedmetadata", function() {
      this.autoplay = true;

      track = document.createElement("track");
      track.kind = "captions";
      track.label = "English";
      track.srclang = "en";
      track.src = jsonResponse.SubUrl;
      track.addEventListener("load", function() {
        this.mode = "showing";
        video.textTracks[0].mode = "showing"; // thanks Firefox
      });
      this.appendChild(track);

      track.addEventListener("error", error => {
        video.removeChild(video.childNodes[0]);
        //console.log(error);
      });
    });

    video.onended = function() {
      var autoplay = document.getElementById("autoplay");
      if(!autoplay.checked) {
        return;
      }

      var episode = document.getElementsByName("episode")[0];
      var serie = document.getElementsByName("serie")[0];
      var season = document.getElementsByName("season")[0];

      var cur_s = season.options[season.selectedIndex].value - 1;
      var cur_e = episode.options[episode.selectedIndex].value;
      var serie = series[serie.options[serie.selectedIndex].value];

      if (cur_e < serie.episodes_max[cur_s]) {
        episode.selectedIndex++;
      } else if(season.selectedIndex + +1 < serie.seasons) {
        season.selectedIndex += +1;
        updateEpisodeCount();
        episode.value = "1";
      }
      playVideo();
    }

    if (videoDiv.childNodes.length > 0) {
      videoDiv.removeChild(videoDiv.childNodes[0]);
    }
    video.setAttribute("style", "margin: auto;");
    videoDiv.appendChild(video);
    video.play();
  });
}


