$(document).ready(function(){
    /*----------------------------------------------------------------
    Submission button handler
    ----------------------------------------------------------------*/
    $("#link_form").submit(function (e) {
        e.preventDefault();

        $("#loading").addClass("spin")
        $("#submit_btn").prop("disabled", true);
        $("#submit_btn").addClass("disabled");
        $("#submit_btn").text("Submitting");

        $.ajax({
            url: "/new_song",
            type: "POST",
            data: $("input"),
            error: function(jqXHR, textStatus, errorThrown) {
                if(jqXHR.status == 500) {
                    $("#alert_area").empty();
                    $("#alert_area").append(jqXHR.responseText);
                } else {
                    alert("Failed to contact server");
                }
                $("#loading").removeClass("spin")
                this.always()
            },
            success: function(data, textStatus, errorThrown) {
                $("#submit_box").val("");
                refresh_elements();
                this.always()
            },
            always: function(jqXHR, textStatus, errorThrown) {
                $("#submit_btn").prop("disabled", false);
                $("#submit_btn").removeClass("disabled");
                $("#submit_btn").text("Submit Link");
            }
        });
    });

    /*----------------------------------------------------------------
    Refresh the things in the queue and now playing banner
    ----------------------------------------------------------------*/
    function refresh_elements() {
        // Start the refresh spinner
        $("#loading").addClass("spin")

        // Make the ajax call to get the now playing song
        $.ajax({
            url: "/now_playing",
            type: "GET",
            dataType: "html",
            error: function(jqXHR, textStatus, errorThrown) {
                if(jqXHR.status == 500) {
                    $("#alert_area").empty();
                    $("#alert_area").append(jqXHR.responseText);
                } else {
                    alert("Failed to contact server")
                }
                $("#loading").removeClass("spin")
            },
            success: function(data, textStatus, errorThrown) {
                $("#banner").empty();
                $("#banner").append(data);
                //disable_wrap();

                // Make the ajax call refresh the queue
                $.ajax({
                    url: "/playlist",
                    type: "GET",
                    dataType: "html",
                    error: function(jqXHR, textStatus, errorThrown) {
                        if(jqXHR.status == 500) {
                            $("#alert_area").empty();
                            $("#alert_area").append(jqXHR.responseText);
                        } else {
                            alert("Failed to contact server")
                        }
                        $("#loading").removeClass("spin")
                    },
                    success: function(data, textStatus, errorThrown) {
                        $("#queue_container").empty();
                        $("#queue_container").append(data);
                        $("#queue_title").click(refresh_elements);
                        $("#queue_title").on("tap", refresh_elements);
                        $(".queue_rm").click(remove_song);
                        $(".skip_now_playing").click(skip_song);
                        flip_collapse_hint();
                    },
                });
            },
        });
    };

    /*----------------------------------------------------------------
    Remove the target song from queue
    ----------------------------------------------------------------*/
    function remove_song(event) {
        $.ajax({
            url: "/remove",
            type: "POST",
            data: { 'song_id' : event.currentTarget.id },
            error: function(jqXHR, textStatus, errorThrown) {
                if(jqXHR.status == 500) {
                    $("#alert_area").empty();
                    $("#alert_area").append(jqXHR.responseText);
                } else {
                    alert("Failed to contact server");
                }
            },
            success: function(data, textStatus, errorThrown) {
                refresh_elements();
            }
        });
    };

    /*----------------------------------------------------------------
    Skip the currently playing song
    ----------------------------------------------------------------*/
    function skip_song(event) {
        $.ajax({
            url: "/next",
            type: "GET",
            data: { 'song_id' : event.currentTarget.id },
            error: function(jqXHR, textStatus, errorThrown) {
                if(jqXHR.status == 500) {
                    $("#alert_area").empty();
                    $("#alert_area").append(jqXHR.responseText);
                } else {
                    alert("Failed to contact server");
                }
            },
            success: function(data, textStatus, errorThrown) {
                refresh_elements();
            }
        });
    };

    /*----------------------------------------------------------------
    Registers handlers to flip the collapse hint on a song in the
    queue
    ----------------------------------------------------------------*/
    function flip_collapse_hint() {
        // go from arrow down to up when the song details
        // are shown
        $("[id^=vid_]").on('shown.bs.collapse', function(event) {
            var song_hint = $(this).parent().next().children(".glyphicon-collapse-down")
            song_hint.toggleClass("glyphicon-collapse-down")
            song_hint.toggleClass("glyphicon-collapse-up")
        });

        // go from arrow up to down when the songs details
        // are hidden
        $("[id^=vid_]").on('hidden.bs.collapse', function(event) {
            var song_hint = $(this).parent().next().children(".glyphicon-collapse-up")
            song_hint.toggleClass("glyphicon-collapse-up")
            song_hint.toggleClass("glyphicon-collapse-down")
        });
    }

    // Register handler on queue items to remove song
    $(".queue_rm").click(remove_song);

    // Register handler to skip the currently playing song
    $(".skip_now_playing").click(skip_song);

    // Register handler on the queue title to refresh items
    $("#queue_title").click(refresh_elements);
    $("#queue_title").on("tap", refresh_elements);

    flip_collapse_hint();
});
