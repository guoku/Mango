<div>
    <a href="#myModal" role="button" class="btn" data-toggle="modal">添加密码</a>
     
    <!-- Modal -->
    <div id="myModal" class="modal hide fade" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true">
      <div class="modal-header">
        <button type="button" class="close" data-dismiss="modal" aria-hidden="true">×</button>
        <h3 id="myModalLabel">添加新密码</h3>
      </div>
      <div class="modal-body">
        <form method="post" action="/add_pass">
          <fieldset>
            <label>名称</label>
            <input type="text" name="name" placeholder="Type something…">
            <label>帐号</label>
            <input type="text" name="account" placeholder="Type something…">
            <label>密码</label>
            <input type="text" name="pwd" placeholder="Type something…">
            <label>描述</label>
            <input type="text" name="desc" placeholder="Type something…">
            <button type="submit" class="btn">保存</button>
          </fieldset>
        </form> 
      </div>
      <div class="modal-footer">
        <button class="btn" data-dismiss="modal" aria-hidden="true">关闭</button>
      </div>
    </div> 
    <table class="table table-striped table-bordered table-condensed">
        <thead>
            <tr>
                <th> Account </th>
                <th> Password </th>
                <th> Desc </th>
                <th> Operation </th>
            </tr>
        </thead>
        <tbody>
            {{ range .User.PasswordPermissions }}
                <tr>
                    <td>{{ .Password.Account }}</td>
                    <td>{{ .Password.Password }}</td>
                    <td>{{ .Password.Desc }}</td>
                    <td>
                        <a href="/edit_pass" class="btn btn-large btn-primary">Edit</a>
                    </td>
                <tr>
            {{end}}
        </tbody>
    </table>
</div>
<script type="text/javascript" src="/static/js/jquery.min.js"></script>
<script type="text/javascript" src="/static/js/bootstrap.min.js"></script>
