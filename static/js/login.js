// 登录函数
async function login() {
    const username = document.getElementById("username").value;
    const password = document.getElementById("password").value;

    if (!username || !password) {
        alert("请输入用户名和密码！");
        return;
    }

    try {
        const response = await fetch("/login", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ username, password }),
        });

        if (response.ok) {
            const data = await response.json();
            // 确保后端返回的 token 是有效的
            if (data.token) {
                localStorage.setItem("token", data.token);
                console.log("登录成功，Token 已保存:", localStorage.getItem("token"));
                window.location.href = "/dashboard/hadoop"; 
            } else {
                alert("登录失败，未收到 Token！");
            }
        } else {
            alert("登录失败，请检查用户名或密码！");
        }
    } catch (error) {
        console.error("请求失败：", error);
        alert("登录失败，请稍后重试！");
    }
}

// 发送带 token 的请求
async function fetchWithToken(url, options = {}) {
    const token = localStorage.getItem("token");

    if (!token) {
        console.log("未找到认证信息，请重新登录！");
        window.location.href = "/login";  
        return;
    }

    // 合并默认请求头和用户提供的请求头
    options.headers = {
        "Authorization": `Bearer ${token}`,
        "Content-Type": "application/json",
    };

    console.log("当前 token:", token);
    console.log("请求 headers:", options.headers);

    try {
        const response = await fetch(url, options);
        if (response.status === 401) {
            alert("认证失败，请重新登录！");
            localStorage.removeItem("token");
            window.location.href = "/login";
            return;
        }
        if (!response.ok) {
            console.error("请求失败:", response.statusText);
            alert("请求失败，请稍后重试！");
        } else {
            return await response.json();
        }
    } catch (error) {
        console.error("请求出错：", error);
        alert("请求出错，请稍后重试！");
    }
}

// 在 Dashboard 页面加载时调用此函数
async function getDashboardData() {
    const data = await fetchWithToken("/dashboard/hadoop", {
        method: "GET",
    });

    if (data) {
        console.log("Dashboard 数据:", data);
        // 处理获取到的数据
    } else {
        alert("获取数据失败，请稍后重试！");
    }
}

// 页面加载时自动执行
document.addEventListener("DOMContentLoaded", async () => {
    const token = localStorage.getItem("token");
    if (token) {
        await getDashboardData(); 
    } else {
        console.log("未找到 token，跳转到登录页面");
        window.location.href = "/login";  // 如果没有 token，自动跳转到登录
    }
});
