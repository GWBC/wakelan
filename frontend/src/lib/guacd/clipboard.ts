import Guacamole from 'guacamole-common-js'

interface ClipData {
  type: string
  data: string
}

export class GuacdClipboard {
  private client: Guacamole.Client
  private focusProc: any

  constructor(client: Guacamole.Client) {
    this.client = client
  }

  private LocalToRemoteClipboard() {
    this.GetLocalClipboard().then((data) => {
      if (!data || !data.data || data.data.length == 0) {
        return
      }

      const stream = this.client.createClipboardStream(data.type)
      let writer = new Guacamole.StringWriter(stream)

      writer.sendText(data.data)
      writer.sendEnd()
    })
  }

  RemoteToLocalClipboard(stream: Guacamole.InputStream, mimeType: string) {
    if (/^text\//.exec(mimeType)) {
      const reader = new Guacamole.StringReader(stream)
      let data = ''
      reader.ontext = (text) => {
        data += text
      }

      reader.onend = () => {
        this.SetLocalClipboard({ type: mimeType, data: data })
      }
    }
  }

  private async GetLocalClipboard(): Promise<ClipData> {
    try {
      if (navigator.clipboard && navigator.clipboard.readText) {
        return {
          type: 'text/plain',
          data: await navigator.clipboard.readText()
        }
      }
    } catch (err) {
      //console.error(err.name, err.message);
    }

    return { type: '', data: '' }
  }

  private async SetLocalClipboard(data: ClipData) {
    try {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        if (data.type === 'text/plain') {
          await navigator.clipboard.writeText(data.data)
        }
      }
    } catch (err) {
      //console.error(err.name, err.message);
    }
  }

  Install() {
    this.focusProc = (e: Event) => {
      if (e.target === window) {
        this.LocalToRemoteClipboard()
      }
    }

    window.addEventListener('focus', this.focusProc, true)
  }

  UnInstall() {
    window.removeEventListener('focus', this.focusProc)
  }
}
