from PIL import Image
import pathlib

root = pathlib.Path(__file__).parent

logo = root / "logo.png"
res  = root / ".." / "wrappers" / "android" / "app" / "src" / "main" / "res"

sizes = {
    "mipmap-mdpi":     48,
    "mipmap-hdpi":     72,
    "mipmap-xhdpi":    96,
    "mipmap-xxhdpi":  144,
    "mipmap-xxxhdpi": 192,
}

src = Image.open(logo).convert("RGBA")

for folder, px in sizes.items():
    d = res / folder
    d.mkdir(exist_ok=True)
    img = src.resize((px, px), Image.LANCZOS)
    img.save(d / "ic_launcher.png")
    img.save(d / "ic_launcher_round.png")
    print(f"  {folder}: {px}x{px}")

print("Done.")
