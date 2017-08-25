"use strict";

var token;
var tags = [];
var selectedChatId = 0;

function displayPhoto(photo) {
    var block = document.createElement('div');

    block.className = 'thumbnail';
    block.innerHTML =
        '<a class="photo-link" href="/images' + photo.src + '">' +
        '<img alt="Фото недоступно" src="/images' + photo.src + '">' +
        '</a>';

    salvattore.appendElements(document.querySelector('#content'), [block]);
};

function setPhotos(tags, chatId) {
    if (chatId == 0) {
        var data = ""
    } else {
        var data = "chat_id=" + chatId;
    }

    tags.forEach(function(tag) {
        if (data != "")
            data += "&";

        data += "tag=" + tag;
    });

    $.ajax({
            url: '/api/v1/images.get',
            dataType: "json",
            type: "get",
            headers: { "X-User-Token": token },
            data: data
        })
        .done(function(data) {
            $(document).find('#more-button-row').hide();
            $(document).find('.alert').hide();

            document.getElementById('content').innerHTML = '<div class="col-lg-3"></div><div class="col-lg-3"></div><div class="col-lg-3"></div><div class="col-lg-3"></div>';
            data.images.forEach(displayPhoto);

            $('.photo-link').magnificPopup({ type: 'image' });
            $(document).find('#content').show();
        });
}

function setToken() {
    let defer = $.Deferred();

    let url = new URL(window.location.href);
    token = url.searchParams.get("token");

    if (token == null)
        console.log("Token not found")
    else
        console.log("Token", token)

    defer.resolve();
    return defer.promise();
};

function setTags() {
    let defer = $.Deferred();

    if (selectedChatId == 0) {
        var url = "/api/v1/users.tags";
        var data = function(params) { return {} }
    } else {
        var url = "/api/v1/chat.tags";
        var data = function(params) {
            return { "chat_id": selectedChatId };
        }
    }

    $.ajax({
        url: url,
        dataType: 'json',
        type: 'get',
        headers: { 'X-User-Token': token },
    }).then(function(data) {
        let i = 0;

        tags = data["tags"].sort(function(a, b) {
            return a.name.localeCompare(b.name);
        })
        .map(function(tag) {
            return {
                id: (i++).toString(),
                text: tag.name
            };
        });

        defer.resolve();
    });

    return defer.promise();
}

function handleChats(data) {
    let i = 0

    var results = [{
        id: i.toString(),
        text: "All chats"
    }];

    data["chats"]
        .sort(function(a, b) {
            return a.title.localeCompare(b.title);
        })
        .forEach(function(item) {
            results.push({
                id: item.chat_id,
                text: item.title
            });
        });

    return { results: results };
}

$(function() {
    $('#tags_select').select2();

    setToken().then(function() {
        setPhotos([], selectedChatId);

        $.fn.select2.defaults.set("theme", "bootstrap");

        $('#chat_select').select2({
            ajax: {
                url: '/api/v1/chats.get',
                dataType: "json",
                headers: { "X-User-Token": token },
                processResults: handleChats,
            },
            minimumResultsForSearch: Infinity
        });

        setTags().then(function() {
            $('#tags_select').select2({
                data: tags,
                allowClear: true
            });

            $('#tags_select').on('change', function(event) {
                setPhotos(
                    $('#tags_select').select2('data').map(function(tag) {
                        return tag.text
                    }),
                    selectedChatId
                );
            });
        });

        $('#chat_select').on('select2:select', function(event) {
            selectedChatId = event.params.data.id;
            setTags().then(function() {
                $('#tags_select').empty();

                $('#tags_select').select2({
                    data: tags[selectedChatId],
                    allowClear: true
                });
            });

            setPhotos([], selectedChatId)
        });
    });
});
