{{ template "nav.tpl" .}}
<div class = "span5">
    <legend>Password Info</legend>
    <form method="post" action="/edit_pass/{{.Password.Id}}">
        <label>名称</label>
        <input type="text" name="name" value="{{ .Password.Name  }}">
        <label>帐号</label>
        <input type="text" name="account" value="{{ .Password.Account }}">
        <label>密码</label>
        <input type="password" name="password" value="{{ .Password.Password }}">
        <label>描述</label>
        <input type="text" name="desc" value="{{ .Password.Desc }}">
        <br>
        <button type="submit" class="btn">保存</button>
    </form>
</div>
{{ if .CanManage }}
<div class="span6">
    <legend>Users</legend>
    <table class="table table-striped table-bordered table-condensed">
        <thead>
            <tr>
                <th> Name </th>
                <th> Level </th>
            </tr>
        </thead>
        <tbody>
            {{ $passid := .Password.Id }}
            {{ range .PassUsers }}
                <tr>
                    <td>
                        {{ .User.Name }}
                    </td>
                    <td>
                    <form class="form-inline" method="post" action="/edit_pass_permission/{{$passid}}">
                        <input type="hidden" name="edit_user_id" value="{{ .User.Id}}">
                        <label class="radio">
                          <input type="radio" name="user_permissions" id="optionsRadios1" value="0" {{if .CheckPermission 0 }} checked {{ end }} >无权限
                        </label>
                        <label class="radio">
                            <input type="radio" name="user_permissions" id="optionsRadios2" value="1" {{if .CheckPermission 1 }} checked {{ end }}>
                                可读
                        </label>
                        <label class="radio">
                            <input type="radio" name="user_permissions" id="optionsRadios3" value="2" {{if .CheckPermission 2 }} checked {{ end }}>
                                可修改
                        </label>
                        <label class="radio">
                            <input type="radio" name="user_permissions" id="optionsRadios4" value="3" {{if .CheckPermission 3 }} checked {{ end }}>
                                可管理
                        </label>
                        <button type="submit" class="btn">保存</button>
                        </form>
                    </td>
                </tr>
            {{ end }}
        </tbody>
    </table>
<div>
{{ end }}
