import { ElMessage } from 'element-plus'
import router from '@/router'

interface FetchResponse<T> {
    (info: T): void
}

//拉取数据
export async function Fetch<T>(url: string, postData: any, resCallback: FetchResponse<T>) {
    let res = null

    if (postData) {
        const requestOptions: RequestInit = {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include',
            body: JSON.stringify(postData)
        };

        res = fetch(url, requestOptions)
    } else {
        res = fetch(url)
    }

    return res.then(response => {
        if (!response.ok) {
            throw response.statusText
        }

        return response.json()
    }).then(data => {
        if (!data) {
            throw new Error("unknown error")
        }

        if (data.err.length != 0) {
            throw data.err
        }

        resCallback(data.infos)
    }).catch(error => {
        if (error.toString().includes("token")) {
            router.push('/login')
        } else {
            console.log(`URL:${url} ${error.toString()}`)
            ElMessage.error(error.toString())
        }
    })
}

//下载文件
export function DownloadFile(data: Blob, filename: string) {
    const url = URL.createObjectURL(data);

    const link = document.createElement('a');
    link.href = url;
    link.download = filename;

    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);

    URL.revokeObjectURL(url);
}

//全屏
export function FullScreenOrRecover(el: any) {
    if (document.fullscreenElement) {
        document.exitFullscreen()
    } else {
        if (el.requestFullscreen) {
            el.requestFullscreen()
        } else if (el.mozRequestFullScreen) { // Firefox
            el.mozRequestFullScreen()
        } else if (el.webkitRequestFullscreen) { // Chrome, Safari and Opera
            el.webkitRequestFullscreen()
        } else if (el.msRequestFullscreen) { // IE/Edge
            el.msRequestFullscreen()
        }
    }
}

//获取当前字符串时间
export function Now2Str(): string {
    const now = new Date()

    const year = now.getFullYear()
    const month = ('0' + (now.getMonth() + 1)).slice(-2)
    const day = ('0' + now.getDate()).slice(-2)
    const hours = ('0' + now.getHours()).slice(-2)
    const minutes = ('0' + now.getMinutes()).slice(-2)
    const seconds = ('0' + now.getSeconds()).slice(-2)

    return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

//深拷贝对象
export function DeepCopy(obj: any) {
    return JSON.parse(JSON.stringify(obj))
}