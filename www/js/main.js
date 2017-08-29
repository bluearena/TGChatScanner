'use strict';

let tags = [];
let selectedChatId = 0;

function displayPhoto(photo, imagesPrefix) {
    let block = document.createElement('div');

    let tagsString = '';
    for (let i = 0; i < photo.tags.length; ++i) {
        tagsString += `<span class='badge'>${photo.tags[i].name}</span>` 
    }

    block.className = 'thumbnail';
    block.innerHTML =
        `<a class="photo-link" title="${tagsString}" href="${imagesPrefix}${photo.src}">` +
            `<img alt="Фото недоступно" src="/images${photo.src}">` +
        `</a>`;

    salvattore.appendElements(document.querySelector('#content'), [block]);
};

function setPhotos(tags, chatId) {
    let settings = {
        url: '/api/v1/images.get',
        dataType: 'json',
        type: 'get',
        headers: {'X-User-Token': localStorage.token},
        traditional: true,
        data: {
            tag: tags
        }
    };

    if (chatId != 0)
        settings.data.chat_id = chatId;

    $.ajax(settings).done(data => {
        $(document).find('#more-button-row').hide();
        $(document).find('.alert').hide();

        document.getElementById('content').innerHTML = '<div class="col-lg-3"></div><div class="col-lg-3"></div><div class="col-lg-3"></div><div class="col-lg-3"></div>';
        data.images.forEach(photo => displayPhoto(photo, data.images_prefix));

        $('.photo-link').magnificPopup({type: 'image'});
        $(document).find('#content').show();
    });
}

function setToken() {
    let defer = $.Deferred();

    let url = new URL(window.location.href);
    let token = url.searchParams.get('token');

    if (token != null) {
        localStorage.token = token;
        history.replaceState({}, "", url.pathname);
    } else if (localStorage.token == null) {
        $('#token-error').show();
    }

    defer.resolve();
    return defer.promise();
};

function setTags() {
    let defer = $.Deferred();

    let settings = {
        dataType: 'json',
        type: 'get',
        headers: {'X-User-Token': localStorage.token},
    };

    if (selectedChatId == 0) {
        settings.url = '/api/v1/users.tags';
    } else {
        settings.url = '/api/v1/chat.tags';
        settings.data = {'chat_id': selectedChatId};
    }

    $.ajax(settings).then(data => {
        let i = 0;

        tags = data['tags'].sort((a, b) => a.name.localeCompare(b.name))
        .map(tag => ({
            id: (i++).toString(),
            text: tag.name
        }));

        defer.resolve();
    });

    return defer.promise();
}

function handleChats(data) {
    let i = 0

    let results = [{
        id: i.toString(),
        text: 'All chats'
    }];

    data['chats'].sort((a, b) => a.title.localeCompare(b.title))
    .forEach(item => {
        results.push({
            id: item.chat_id,
            text: item.title
        });
    });

    return {results};
}

$(() => {
    $.fn.select2.defaults.set('theme', 'bootstrap');
    $('#tags_select').select2();

    setToken().then(() => {
        setPhotos([], selectedChatId);

        $('#chat_select').select2({
            ajax: {
                url: '/api/v1/chats.get',
                dataType: 'json',
                headers: {'X-User-Token': localStorage.token},
                processResults: handleChats,
            },
            minimumResultsForSearch: Infinity
        });

        setTags().then(() => {
            $('#tags_select').select2({
                data: tags,
                allowClear: true
            });

            $('#tags_select').on('change', event => {
                setPhotos(
                    $('#tags_select').select2('data').map(tag => tag.text),
                    selectedChatId
                );
            });
        });

        $('#chat_select').on('select2:select', event => {
            selectedChatId = event.params.data.id;

            setTags().then(() => {
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
