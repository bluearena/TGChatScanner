"use strict";

var token;
var tags = [];
var selectedChatId = 0;

function displayPhoto(photo) {
    let block = document.createElement('div');

    let tags = "";
    photo.tags.forEach(function(tag) {
        tags += '<span class=\'badge\'>' + tag.name + '</span>' 
    });

    block.className = 'thumbnail';
    block.innerHTML =
        '<a class="photo-link" title="' + tags + '" href="/images' + photo.src + '">' +
            '<img alt="Фото недоступно" src="/images' + photo.src + '">' +
        '</a>';

    salvattore.appendElements(document.querySelector('#content'), [block]);
};

function setPhotos(tags, chatId) {
    let data;
    if (chatId == 0) {
        data = ""
    } else {
        data = "chat_id=" + chatId;
    }

    tags.forEach(function(tag) {
        if (data != "")
            data += "&";

        data += "tag=" + tag;
    });

    $.ajax({
        url: '/api/v1/images.get',
        dataType: 'json',
        type: 'get',
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
        $('#token-error').show();

    defer.resolve();
    return defer.promise();
};

function setTags() {
    let defer = $.Deferred();
    let url, data;

    if (selectedChatId == 0) {
        url = '/api/v1/users.tags';
        data = {};
    } else {
        url = '/api/v1/chat.tags';
        data = {'chat_id': selectedChatId};
    }

    $.ajax({
        url: url,
        dataType: 'json',
        type: 'get',
        headers: { 'X-User-Token': token },
        data: data
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

    let results = [{
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
                headers: { "X-User-Token": token },
                dataType: 'json',
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
                    data: tags,
                    allowClear: true
                });
            });

            setPhotos([], selectedChatId)
        });
    });
});
