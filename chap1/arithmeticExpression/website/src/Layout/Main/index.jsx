import { Box, Typography, Stack, TextField, Grid, Button } from '@mui/material';
import { useState } from 'react';
import { cal } from '../../api/calculator';
const Main = () => {
    const [expression, setExpression] = useState("")
    const [result, setResult] = useState("")
    const op = ["%", "CE", "C", "del", "1/x", "x^2", "x^(1/2)", "÷", "7", "8", "9", "X", "4", "5", "6", "-", "1", "2", "3", "+", "+/-", "0", ".", "="]
    const change = (event) => {
        setExpression(event.target.value)
    }
    const click = async (e) => {
        let v = e.currentTarget.getAttribute("value")
        switch (v) {
            case "X":
                v = "*"
                break;
            case "÷":
                v = "/"
            case "=":
                let res = await cal({ expression })
                setResult(res.data)
                return
            default:
                break;
        }
        setExpression(expression + v)
    }
    return (
        <Stack sx={{ width: "400px", height: "600px" }} justifyContent={"center"} alignItems={"center"} direction={"column"}>
            <Typography variant='h1'>计算器</Typography>
            <Box >
                <Typography variant='h1' sx={{ width: "400px", height: "100px", textAlign: "right" }}>{result}</Typography>
                <Typography variant='h1' sx={{ width: "400px", height: "100px", textAlign: "right" }}>{expression}</Typography>
            </Box>
            <Grid container spacing={1}>
                {
                    op.map((v, i) => {
                        return <Grid item xs={3}>
                            <Button fullWidth variant={v === '=' ? "contained" : "outlined"} value={v} onClick={click}>{v}</Button>
                        </Grid>
                    })
                }
            </Grid>
        </Stack >
    )
}
export default Main