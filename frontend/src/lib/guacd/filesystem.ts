import Guacamole from 'guacamole-common-js'
import { DownloadFile } from '@/lib/comm'

export interface SFTPFileInfo {
    name: string
    path: string
    type: string
}

type DropFun = (filename: string, data: Blob, type: string) => void
type FilelistFun = (filelist: SFTPFileInfo[]) => void
type UPProcFun = (size: number, offset: number, err: any) => void
type DownProcFun = (recvSize: number, isEnd: boolean, err: any) => void

export class GuacdFilesystem {
    private curPath: string = '/'
    private dragEl!: Element
    private dropFun!: DropFun
    private filesystemObj!: Guacamole.Object

    constructor() {
    }

    private onDragover(ev: Event) {
        ev.preventDefault()
    }

    private onDragleave(ev: Event) {

    }

    private onDrop(ev: Event) {
        let e = ev as DragEvent

        e.preventDefault();

        if (!e.dataTransfer) {
            return
        }

        // 获取拖拽的文件
        let files = e.dataTransfer.files;
        for (let i = 0; i < files.length; ++i) {
            const file = files[i]
            const reader = new FileReader()
            reader.onloadend = (freader) => {
                if (!freader.target || !freader.target.result) {
                    return
                }

                if (this.dropFun) {
                    let data = new Blob([freader.target.result], { type: file.type })
                    this.dropFun(file.name, data, file.type)
                }
            }

            reader.readAsArrayBuffer(file)
        }
    }

    //安装
    Install(dragEl: HTMLElement, onDrop: DropFun): void {
        this.dragEl = dragEl
        this.dropFun = onDrop
        this.dragEl.addEventListener('dragover', this.onDragover);
        this.dragEl.addEventListener('dragleave', this.onDragleave);
        this.dragEl.addEventListener('drop', this.onDrop);
    }

    //卸载
    UnInstall() {
        if (!this.dragEl) {
            return
        }

        this.dragEl.removeEventListener('dragover', this.onDragover);
        this.dragEl.removeEventListener('dragleave', this.onDragleave);
        this.dragEl.removeEventListener('drop', this.onDrop);
    }

    //SFTP文件系统接入
    onFileSystem(obj: Guacamole.Object, name: string) {
        this.filesystemObj = obj
    }

    //获取当前路径
    CurPath(): string {
        return this.curPath
    }

    //切换目录
    CD(path: string, onFilelist: FilelistFun): boolean {
        return this.DirList(path, (fileList) => {
            this.curPath = path

            if (onFilelist) {
                onFilelist(fileList)
            }
        })
    }

    //获取文件列表
    DirList(path: string, onFilelist: FilelistFun): boolean {
        if (!this.filesystemObj) {
            return false
        }

        this.filesystemObj.requestInputStream(path, (stream, mimeType) => {
            if (mimeType != Guacamole.Object.STREAM_INDEX_MIMETYPE) {
                let fileList = [{
                    name: path,
                    path: path,
                    type: "file"
                }]

                let index = path.lastIndexOf("/")
                if (index >= 0) {
                    fileList[0].name = path.substring(index + 1)
                }

                if (onFilelist) {
                    onFilelist(fileList)
                }

                stream.sendAck('Unexpected mimetype', Guacamole.Status.Code.UNSUPPORTED);
                return;
            }

            stream.sendAck('Ready', Guacamole.Status.Code.SUCCESS);

            let reader = new Guacamole.JSONReader(stream);

            reader.onprogress = () => {
                stream.sendAck("Received", Guacamole.Status.Code.SUCCESS);
            };

            reader.onend = () => {
                let fileList: SFTPFileInfo[] = []
                var mimetypes = reader.getJSON() as Record<string, string | number>

                for (let filepath in mimetypes) {
                    if (filepath.substring(0, path.length) !== path) {
                        continue;
                    }

                    let filename = filepath.substring(path.length);
                    if (path.substring(path.length - 1) != '/') {
                        filename = filepath.substring(path.length + 1);
                    }

                    fileList.push({
                        name: filename,
                        path: filepath,
                        type: mimetypes[filepath] == Guacamole.Object.STREAM_INDEX_MIMETYPE ? "dir" : "file"
                    })
                }

                if (onFilelist) {
                    onFilelist(fileList)
                }
            }
        })

        return true
    }

    //下载
    Down(path: string, onDownProc: DownProcFun) {
        this.DirList(path, (fileList) => {
            for (let i = 0; i < fileList.length; ++i) {
                let fileInfo = fileList[i]

                //处理目录
                if (fileInfo.type == 'dir') {
                    this.Down(fileInfo.path, onDownProc)
                    continue
                }

                //处理文件
                this.filesystemObj.requestInputStream(fileInfo.path, (stream, mimeType) => {
                    try {
                        if (mimeType == Guacamole.Object.STREAM_INDEX_MIMETYPE) {
                            stream.sendAck('Unexpected mimetype', Guacamole.Status.Code.UNSUPPORTED);
                            return
                        }

                        stream.sendAck('Ready', Guacamole.Status.Code.SUCCESS);

                        let reader = new Guacamole.BlobReader(stream, mimeType);

                        reader.onprogress = (length) => {
                            if (onDownProc) {
                                onDownProc(reader.getLength(), false, '')
                            }

                            stream.sendAck("Received", Guacamole.Status.Code.SUCCESS);
                        };

                        reader.onend = () => {
                            if (onDownProc) {
                                onDownProc(reader.getLength(), true, '')
                            }

                            DownloadFile(reader.getBlob(), fileInfo.name)
                        }
                    } catch (err: any) {
                        if (onDownProc) {
                            onDownProc(0, true, err)
                        }
                    }
                })
            }
        })
    }


    //上传
    Upload(data: Blob, mimeType: string, filepath: string, onUPProc: UPProcFun): boolean {
        if (!this.filesystemObj) {
            return false
        }

        try {
            let stream = this.filesystemObj.createOutputStream(mimeType, filepath)

            let writer = new Guacamole.BlobWriter(stream);

            writer.onerror = (blob, offset, err) => {
                if (onUPProc) {
                    onUPProc(0, 0, err)
                }

                writer.sendEnd()
            }

            writer.onprogress = (blob, offset) => {
                if (onUPProc) {
                    onUPProc(blob.size, offset, '')
                }
            }

            writer.oncomplete = (blob) => {
                if (onUPProc) {
                    onUPProc(blob.size, blob.size, '')
                }
                writer.sendEnd()
            }

            writer.sendBlob(data)
        } catch (err: any) {
            if (onUPProc) {
                onUPProc(0, 0, err)
            }
        }

        return true
    }
}

