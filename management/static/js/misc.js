/**
 * Created with PyCharm.
 * User: edison
 * Date: 12-8-9
 * Time: PM1:44
 * To change this template use File | Settings | File Templates.
 */

$(document).ready(function() {
    initNavForm();
    initDeleteTopicMap();
    initDeleteTopic();
    initDeleteFolder();
    initDeleteComment();
    initAddCatalog();
    initDeleteCatalog();
//    initDeleteCategory();
    initImageSelect();
    initPubCandidate();
    vDateClass();
    vTimeClass();

    actionAddNewVer();
    actionSaveNewVersion();
    $(function(){
        $('#dp1').datepicker({
            format: 'yyyy-mm-dd'
        });
    });
});

function initNavForm(){
    var form = $(".navbar-search");
    form.submit(function(){
        var form_name = this.form_name.value;
        this.action = "/management/"+ form_name + "/list/"
    });
}

function initDeleteTopicMap(){
    var _topics = jQuery('div.topics a');
    _topics.each(function() {
            var _topic = $(this);
        _topic.click(function(){
            var _target = $(this).attr('href');
            var _tid = $(this).attr('id');
            var _fid = $(this).attr('fid');
            $.ajax({
                type:"POST",
                url: _target,
                data: {tid:_tid, fid:_fid},
                success:function(msg){
//                    $(this).hidden();
                    _topic.remove();
                }
            });
            return false;
        });
    });
}

function initDeleteTopic() {
    var  _topics = jQuery('tbody.topic-list tr');
//    console.log(_topics);
    _topics.each(function(){
        var _topic = $(this);
        var _link = _topic.find('a.deltopic');
        var _tid = _topic.attr('id');
//        console.log(_tid);
        _link.click(function(){
            var _target = $(this).attr('href');
            $.ajax({
                type:"POST",
                url:_target,
                data:{'tid':_tid},
                success: function(msg) {
                    _topic.remove();
                }
            })
            return false;
        })
    });
}

function initDeleteFolder() {
    var _folders = jQuery('tbody.folder-list tr');

    _folders.each(function(){
        var _folder = $(this);
        var _link = _folder.find('a.delfolder');
        var _fid = _folder.attr('id');
//        console.log(_fid);
        _link.click(function() {
            var _target = $(this).attr('href');
            $.ajax({
                type:"POST",
                url:_target,
                data:{'fid':_fid},
                success: function(msg){
                    _folder.remove();
                }
            })
            return false;
        })
    })
}

function initDeleteComment() {
    var _comments = jQuery('tbody.comment-list tr');
//    console.log(_comment);
    _comments.each(function(){
        var _comment = $(this);
        var _link = _comment.find('a.delcomment');
        var _cid = _comment.attr('id');
        _link.click(function(){
            var _target = $(this).attr('href');
            $.ajax({
                type:"POST",
                url:_target,
                data:{'cid':_cid},
                success: function(msg){
                    _comment.remove();
                }
            })
            return false;
        })
    })
}

function initAddCatalog() {
    var _catalog_table = jQuery('tbody.catalog-list');
//    console.log(_catalog_table);
    var _catalog = jQuery('#add-catalog');
    var _form = jQuery('#catalog-form');

    var _save = _catalog.find('a.addcatalog');
    _save.click(function(){
        var _target = $(this).attr('href');
        $.ajax({
            type:'POST',
            url: _target,
            data:_form.serialize(),
            success: function(msg) {
                _catalog.modal('hide');
                location.reload();
            }
        });
        return false;
    })
}

function initDeleteCatalog() {
    var _catalogs = jQuery('tbody.catalog-list tr');
//    console.log(_catalogs);
    _catalogs.each(function(){
        var _catalog = $(this);
        var _link = _catalog.find('a.delete-catalog')
        var _cid = _catalog.attr('id');
//        console.log(_link);
        _link.click(function(){
            var _target = $(this).attr('href');
//            console.log(_cid);
            $.ajax({
                type:'POST',
                url:_target,
                data:{'cid':_cid},
                success: function(msg) {
                    _catalog.remove();
                }
            })
            return false;
        })
    })
}

function initImageSelect() {
    var _thumbnails = jQuery('#chlid-images ul');
    var _images = _thumbnails.find("li a");
//    console.log(_images)
    _images.each(function(){
        var _image = $(this);
        _image.click(function(){
            var _img_link = $(this).find('img').attr('src');
            var _update_link = jQuery('#img_url');
            var _main_image_link = jQuery('#main-image');
            _update_link.val(_img_link);
            _main_image_link.attr('src', _img_link);
//            console.log(_img_link);
            return false;
        })
    });
}

function initPubCandidate(){
    var _unpass = 'label-warning';
    var _pass = 'label-success';
    var _pub_menus = jQuery('.dropdown-menu');
    _pub_menus.each(function(){
        var _pub_menu = $(this);
        var _links = _pub_menu.find('a');
        _links.each(function(){
            var _link = $(this);
            _link.click(function(){
                var _target = $(this).attr('href');
                var _method = $(this).attr('hreflang');
                $.ajax({
                    type:'POST',
                    url:_target,
                    data: {'method':_method},
                    success: function(msg){
                        if (msg === "1"){
                            var arrs = _target.split("/");
                            var _list_id = arrs[5];
                            var _status = jQuery('#status-'+_list_id);
                            var _status_val = _status.find('span');
                            var _operation_btn = jQuery('#operation-'+_list_id);
                            _status_val.removeClass(_unpass);
                            _status_val.addClass(_pass);
                            _status_val.html('已通过');
                            _operation_btn.remove();
                            _pub_menu.remove();
                        }
                    }
                });
                return false;
            });
        })
    });
}

function vDateClass(){
    var _dateControl = jQuery('div .vDateClass');
    var _inputDate = _dateControl.find('input');
    var _link = _dateControl.find('a');
    _link.click(function(){
        var now = new Date();
//        now.format()
//        var df = new dateFormat();
//        var _dateString = now.getFullYear() + '-' + ((now.getMonth() + 1) + '-' + now.getDate();
        var _dateString = now.format('yyyy-mm-dd');
        _inputDate.val(_dateString);
        return false;
    })
}

function vTimeClass(){
    var _dateControl = jQuery('div .vTimeClass');
    var _inputDate = _dateControl.find('input');
    var _link = _dateControl.find('a');
    _link.click(function(){
        var now = new Date();
        var _timeString = now.format('HH:MM:ss');
//        var _timeString = now.getHours() + ':' + now.getMinutes() + ':' + now.getUTCSeconds();
        _inputDate.val(_timeString);
        return false;
    })
}

function actionAddNewVer(){
    var _addNewVersion = $('.add-action');
    var _modalView = $('#guokuModal');
    _addNewVersion.click(function(){
        _modalView.modal({
            show:true
        });
    });
//    _modalView.on('shown', function(){
////        console.log($(this));
//
//    });
}

function actionSaveNewVersion(){
    var _ver_table = $('table tbody');
//    console.log(_ver_table);
    var _modalView = $('#guokuModal');
        var _form = _modalView.find('.add-form');
//        console.log(_form.serialize());
        var _save = _modalView.find('.save');
        _save.click(function(){
            var _link = _form.attr('action');
            $.ajax({
                type:"POST",
                url:_link,
                data:_form.serialize(),
                success: function(msg){
                    _ver_table.prepend(msg);
//                    console.log(msg);
                    _modalView.modal('hide');
                }
            })
        });
}
