{{ template "nav.tpl" }}

<div>
    <div class="modal-header">
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
                <input type="password" name="password" placeholder="Type something…">
                <label>描述</label>
                <input type="text" name="desc" placeholder="Type something…">
                <br>
                <button type="submit" class="btn">保存</button>
            </fieldset>
        </form> 
    </div>
</div> 

