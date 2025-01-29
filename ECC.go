package main

import (
	"bufio"
	"crypto/elliptic"
	"fmt"
	"image/color"
	"log"
	"math/big"
	"os"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func main() {
	// 楕円曲線を選択 (P256)
	curve := elliptic.P256()

	// バッファ付きリーダーを作成
	reader := bufio.NewReader(os.Stdin)

	// 公開鍵をユーザーから取得
	fmt.Println("公開鍵の X 座標を入力してください (16進数形式):")
	xStr, _ := reader.ReadString('\n')
	xStr = strings.TrimSpace(xStr)

	fmt.Println("公開鍵の Y 座標を入力してください (16進数形式):")
	yStr, _ := reader.ReadString('\n')
	yStr = strings.TrimSpace(yStr)

	// 16進数文字列を big.Int に変換
	x, ok := new(big.Int).SetString(xStr, 16)
	if !ok {
		log.Fatalf("X 座標の変換に失敗しました: %s", xStr)
	}
	y, ok := new(big.Int).SetString(yStr, 16)
	if !ok {
		log.Fatalf("Y 座標の変換に失敗しました: %s", yStr)
	}

	// 入力された公開鍵が曲線上にあるか確認
	if !curve.IsOnCurve(x, y) {
		log.Fatalf("公開鍵が楕円曲線上にありません")
	}
	fmt.Println("公開鍵は楕円曲線上にあります")

	// プロットの設定
	p := plot.New()
	p.Title.Text = "ECC Curve and Public Key"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	// 表示範囲の設定
	margin := float64(100)
	xFloat := float64(x.Int64())
	yFloat := float64(y.Int64())
	p.X.Min = xFloat - margin
	p.X.Max = xFloat + margin
	p.Y.Min = yFloat - margin
	p.Y.Max = yFloat + margin

	// 楕円曲線のポイントをプロット
	pts := generateCurvePoints(curve, p.X.Min, p.X.Max, 1000)
	scatter, err := plotter.NewScatter(pts)
	if err != nil {
		log.Fatalf("グラフデータの作成に失敗しました: %v", err)
	}
	scatter.Radius = vg.Points(1)

	// 公開鍵の点を追加
	keyPts := plotter.XYs{{X: xFloat, Y: yFloat}}
	keyPoint, err := plotter.NewScatter(keyPts)
	if err != nil {
		log.Fatalf("公開鍵のプロットに失敗しました: %v", err)
	}
	keyPoint.Radius = vg.Points(3)
	keyPoint.Color = color.RGBA{R: 255, B: 0, G: 0, A: 255} // 赤色を使用

	// グリッドの追加
	p.Add(plotter.NewGrid())

	// グラフに追加
	p.Add(scatter, keyPoint)
	p.Legend.Add("Curve Points", scatter)
	p.Legend.Add("Public Key", keyPoint)

	// グラフをファイルに保存
	if err := p.Save(8*vg.Inch, 8*vg.Inch, "ecc_curve.png"); err != nil {
		log.Fatalf("グラフの保存に失敗しました: %v", err)
	}
	fmt.Println("グラフが ecc_curve.png に保存されました")
}

// 楕円曲線上の点を生成
func generateCurvePoints(curve elliptic.Curve, xMin, xMax float64, numPoints int) plotter.XYs {
	var points plotter.XYs
	p := curve.Params()

	step := (xMax - xMin) / float64(numPoints)
	for x := xMin; x <= xMax; x += step {
		if x < 0 {
			continue
		}
		xBig := big.NewInt(int64(x))
		if xBig.Cmp(p.P) >= 0 {
			continue
		}

		// y² = x³ + ax + b mod p
		y2 := new(big.Int).Exp(xBig, big.NewInt(3), p.P)
		y2.Add(y2, new(big.Int).Mul(xBig, p.B))
		y2.Add(y2, p.B)
		y2.Mod(y2, p.P)

		yBig := new(big.Int).ModSqrt(y2, p.P)
		if yBig != nil {
			points = append(points, plotter.XY{X: x, Y: float64(yBig.Int64())})
			negY := new(big.Int).Sub(p.P, yBig)
			points = append(points, plotter.XY{X: x, Y: float64(negY.Int64())})
		}
	}
	return points
}
