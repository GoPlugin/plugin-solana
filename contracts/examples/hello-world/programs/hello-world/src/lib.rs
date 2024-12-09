use anchor_lang::prelude::*;

use plugin_solana as plugin;

declare_id!("Fg6PaFpoGXkYsidMpWTK6W2BeZ7FEfcYkg476zPFsLnS");

struct Decimal {
    pub value: i128,
    pub decimals: u32,
}

impl Decimal {
    pub fn new(value: i128, decimals: u32) -> Self {
        Decimal { value, decimals }
    }
}

impl std::fmt::Display for Decimal {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let mut scaled_val = self.value.to_string();
        if scaled_val.len() <= self.decimals as usize {
            scaled_val.insert_str(
                0,
                &vec!["0"; self.decimals as usize - scaled_val.len()].join(""),
            );
            scaled_val.insert_str(0, "0.");
        } else {
            scaled_val.insert(scaled_val.len() - self.decimals as usize, '.');
        }
        f.write_str(&scaled_val)
    }
}

#[program]
pub mod hello_world {
    use super::*;
    pub fn execute(ctx: Context<Execute>) -> Result<()> {
        let round = plugin::latest_round_data(
            ctx.accounts.plugin_program.to_account_info(),
            ctx.accounts.plugin_feed.to_account_info(),
        )?;

        let description = plugin::description(
            ctx.accounts.plugin_program.to_account_info(),
            ctx.accounts.plugin_feed.to_account_info(),
        )?;

        let decimals = plugin::decimals(
            ctx.accounts.plugin_program.to_account_info(),
            ctx.accounts.plugin_feed.to_account_info(),
        )?;

        let decimal = Decimal::new(round.answer, u32::from(decimals));
        msg!("{} price is {}", description, decimal);
        Ok(())
    }
}

#[derive(Accounts)]
pub struct Execute<'info> {
    /// CHECK:
    pub plugin_feed: AccountInfo<'info>,
    /// CHECK: TODO: add Anchor types to plugin-solana
    pub plugin_program: AccountInfo<'info>,
}
