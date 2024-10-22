# jfsh

A terminal-based user interface for [Jellyfin](https://jellyfin.org) that lets you browse your media library and play videos directly from the command line.

There's a better project called [jftui](https://github.com/Aanok/jftui) that I 'borrowed' the `libmpv` code from, it's guaranted to work better than this. I'd use it instead of this if I were you.
The reason I had for making this is because `jftui` didn't work for me on Mac and I don't know C enough to fix it and I wanted something that works on Linux _and_ Mac, and also the fact that it's called `jftui` but it doesn't really seem like a TUI in my opinion, it's more of a 'shell'.
So I decided to make a new project and hopefully make something that's more TUIsh and more working on Mac so I made it and called it `jfsh`.

But `jfsh` is currently only on Linux and it works for the most part. I couldn't get `libmpv` to work on Mac. I'm thinking of ditching `libmpv` all together and just running `mpv` and using sockets to communicate with it. Which would drop the requirement for `libmpv` and make it run on everything.

## Warning

This is only working on Linux right now and there are probably a lot of bugs. Make a pull request or an issue.

## Installation

### Prerequisites

- A running Jellyfin instance.
- `libmpv` which comes with `mpv` as far as I know.

### Install via go

```sh
go install github.com/hacel/jfsh@latest
```

## Usage

1. **Start jfsh**

   ```sh
   jfsh
   ```

2. **Login**

   On first launch, you'll be prompted to enter:

   - **Host**: e.g., `http://localhost:8096`
   - **Username**
   - **Password**

3. **Navigate**

   - Use the **arrow keys** or **`hjkl``** to move through menus.

4. **Play Media**

   - Select an item and press **Enter** or **Space** to play it.
   - `mpv` will launch and begin streaming.

5. **Quit**

   - Press **`q`** at any time to exit jfsh.

## Configuration

Configuration files are stored in `~/.config/jfsh/jfsh.yaml`, there's not really any configuration yet. That's just where the secret variables are stored.

## TODO

- Darwin support
- Search
