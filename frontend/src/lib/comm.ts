import { ElMessage } from 'element-plus'
import router from '@/router'
import CryptoJS from 'crypto-js'

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

        try {
            return response.json()
        } catch (errors) {
            throw errors
        }
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
    DownloadFileFromURL(url, filename)
    URL.revokeObjectURL(url);
}

//下载文件
export function DownloadFileFromURL(url: string, filename: string) {
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;

    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
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

export function AESEncrypt(msg: string, key: string, iv: string): string {
    let data = CryptoJS.AES.encrypt(CryptoJS.enc.Utf8.parse(msg),
        CryptoJS.enc.Utf8.parse(key), {
        iv: CryptoJS.enc.Utf8.parse(iv),
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.ZeroPadding
    })

    return data.ciphertext.toString(CryptoJS.enc.Base64url)
}

export function AESDecrypt(msg: string, key: string, iv: string): string {
    let data = CryptoJS.lib.CipherParams.create({
        ciphertext: CryptoJS.enc.Base64url.parse(msg)
    });

    let res = CryptoJS.AES.decrypt(data,
        CryptoJS.enc.Utf8.parse(key), {
        iv: CryptoJS.enc.Utf8.parse(iv),
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.ZeroPadding
    })

    return res.toString(CryptoJS.enc.Utf8)
}

export function DeleteCookie(key: string) {
    document.cookie = `${key}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/api;`
}
